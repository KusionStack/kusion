package operation

import (
	"sync"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation/graph"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/operation/parser"
	"kusionstack.io/kusion/pkg/engine/release"
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/third_party/terraform/dag"
	"kusionstack.io/kusion/third_party/terraform/tfdiags"
)

type DestroyOperation struct {
	models.Operation
}

type DestroyRequest struct {
	models.Request
	Release *apiv1.Release
}

type DestroyResponse struct {
	Release *apiv1.Release
}

// Destroy will delete all resources in this request. The whole process is similar to the operation Apply,
// but every node's execution is deleting the resource.
func (do *DestroyOperation) Destroy(req *DestroyRequest) (rsp *DestroyResponse, s v1.Status) {
	o := do.Operation
	defer close(o.MsgCh)

	if s = validateDestroyRequest(req); v1.IsErr(s) {
		return nil, s
	}

	// 1. init & build Indexes
	priorState := req.Release.State
	priorStateResourceIndex := priorState.Resources.Index()
	// copy priorStateResourceIndex into a new map
	stateResourceIndex := map[string]*apiv1.Resource{}
	for k, v := range priorStateResourceIndex {
		stateResourceIndex[k] = v
	}

	// only destroy resources we have recorded
	resources := priorState.Resources
	runtimesMap, s := runtimeinit.Runtimes(*req.Release.Spec)
	if v1.IsErr(s) {
		return nil, s
	}
	o.RuntimeMap = runtimesMap

	// 2. build & walk DAG
	destroyGraph, s := newDestroyGraph(resources)
	if v1.IsErr(s) {
		return nil, s
	}

	rel, s := copyRelease(req.Release)
	if v1.IsErr(s) {
		return nil, s
	}
	destroyOperation := &DestroyOperation{
		Operation: models.Operation{
			OperationType:           models.Destroy,
			ReleaseStorage:          o.ReleaseStorage,
			CtxResourceIndex:        map[string]*apiv1.Resource{},
			PriorStateResourceIndex: priorStateResourceIndex,
			StateResourceIndex:      stateResourceIndex,
			RuntimeMap:              o.RuntimeMap,
			Stack:                   o.Stack,
			MsgCh:                   o.MsgCh,
			Lock:                    &sync.Mutex{},
			Release:                 rel,
			Sem:                     o.Sem,
		},
	}

	w := &dag.Walker{Callback: destroyOperation.walkFun}
	w.Update(destroyGraph)
	// Wait
	if diags := w.Wait(); diags.HasErrors() {
		s = v1.NewErrorStatus(diags.Err())
		return nil, s
	}

	return &DestroyResponse{Release: destroyOperation.Release}, nil
}

func (do *DestroyOperation) walkFun(v dag.Vertex) (diags tfdiags.Diagnostics) {
	return applyWalkFun(&do.Operation, v)
}

func validateDestroyRequest(req *DestroyRequest) v1.Status {
	if req == nil {
		return v1.NewErrorStatusWithMsg(v1.InvalidArgument, "request is nil")
	}
	if err := release.ValidateRelease(req.Release); err != nil {
		return v1.NewErrorStatusWithMsg(v1.InvalidArgument, err.Error())
	}
	if req.Release.Phase != apiv1.ReleasePhaseDestroying {
		return v1.NewErrorStatusWithMsg(v1.InvalidArgument, "release phase is not destroying")
	}
	return nil
}

func newDestroyGraph(resource apiv1.Resources) (*dag.AcyclicGraph, v1.Status) {
	ag := &dag.AcyclicGraph{}
	ag.Add(&graph.RootNode{})
	deleteResourceParser := parser.NewDeleteResourceParser(resource)
	status := deleteResourceParser.Parse(ag)
	if v1.IsErr(status) {
		return nil, status
	}

	return ag, status
}

package operation

import (
	"errors"
	"fmt"
	"sync"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation/graph"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/release"
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/third_party/terraform/dag"
	"kusionstack.io/kusion/third_party/terraform/tfdiags"
)

type PreviewOperation struct {
	models.Operation
}

type PreviewRequest struct {
	models.Request
	Spec  *apiv1.Spec
	State *apiv1.State
}

type PreviewResponse struct {
	Order *models.ChangeOrder
}

// Preview compute all changes between resources in request and the actual infrastructure.
// The whole process is similar to the operation Apply, but the execution of each node is mocked and will not actually invoke the Runtime
func (po *PreviewOperation) Preview(req *PreviewRequest) (rsp *PreviewResponse, s v1.Status) {
	o := po.Operation

	defer func() {
		if e := recover(); e != nil {
			log.Error("preview panic:%v", e)

			switch x := e.(type) {
			case string:
				s = v1.NewErrorStatus(fmt.Errorf("preview panic:%s", e))
			case error:
				s = v1.NewErrorStatus(x)
			default:
				s = v1.NewErrorStatus(errors.New("unknown panic"))
			}
		}
	}()

	if s = validatePreviewRequest(req); v1.IsErr(s) {
		return nil, s
	}

	// 1. init & build Indexes
	priorState := req.State

	// Kusion is a multi-runtime system. We initialize runtimes dynamically by resource types
	runtimesMap, s := runtimeinit.Runtimes(*req.Spec)
	if v1.IsErr(s) {
		return nil, s
	}
	o.RuntimeMap = runtimesMap

	var (
		priorStateResourceIndex map[string]*apiv1.Resource
		ag                      *dag.AcyclicGraph
	)
	switch o.OperationType {
	case models.ApplyPreview:
		priorStateResourceIndex = priorState.Resources.Index()
		ag, s = newApplyGraph(req.Spec, priorState)
	case models.DestroyPreview:
		resources := req.Spec.Resources
		priorStateResourceIndex = resources.Index()
		ag, s = newDestroyGraph(resources)
	}
	if v1.IsErr(s) {
		return nil, s
	}
	// copy priorStateResourceIndex into a new map
	stateResourceIndex := map[string]*apiv1.Resource{}
	for k, v := range priorStateResourceIndex {
		stateResourceIndex[k] = v
	}

	// 2. walk DAG and preview resources
	log.Info("walking DAG and preview resources ...")

	previewOperation := &PreviewOperation{
		Operation: models.Operation{
			OperationType:           o.OperationType,
			ReleaseStorage:          o.ReleaseStorage,
			SecretStore:             req.Spec.SecretStore,
			CtxResourceIndex:        map[string]*apiv1.Resource{},
			PriorStateResourceIndex: priorStateResourceIndex,
			StateResourceIndex:      stateResourceIndex,
			IgnoreFields:            o.IgnoreFields,
			ChangeOrder:             o.ChangeOrder,
			RuntimeMap:              o.RuntimeMap,
			Stack:                   o.Stack,
			Lock:                    &sync.Mutex{},
			Sem:                     o.Sem,
		},
	}

	w := &dag.Walker{Callback: previewOperation.previewWalkFun}
	w.Update(ag)
	// Wait
	if diags := w.Wait(); diags.HasErrors() {
		return nil, v1.NewErrorStatus(diags.Err())
	}

	return &PreviewResponse{Order: previewOperation.ChangeOrder}, nil
}

func (po *PreviewOperation) previewWalkFun(v dag.Vertex) (diags tfdiags.Diagnostics) {
	var s v1.Status
	if v == nil {
		return nil
	}

	po.Sem.Acquire()
	defer po.Sem.Release()

	if node, ok := v.(graph.ExecutableNode); ok {
		s = node.Execute(&po.Operation)
		if v1.IsErr(s) {
			diags = diags.Append(fmt.Errorf("preview failed.\n%v", s))
			return diags
		}
	}
	return nil
}

func validatePreviewRequest(req *PreviewRequest) v1.Status {
	if req == nil {
		return v1.NewErrorStatusWithMsg(v1.InvalidArgument, "request is nil")
	}
	if err := release.ValidateSpec(req.Spec); err != nil {
		return v1.NewErrorStatusWithMsg(v1.InvalidArgument, err.Error())
	}
	return nil
}

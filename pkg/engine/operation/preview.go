package operation

import (
	"errors"
	"fmt"
	"sync"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation/graph"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/third_party/terraform/dag"
	"kusionstack.io/kusion/third_party/terraform/tfdiags"
)

type PreviewOperation struct {
	opsmodels.Operation
}

type PreviewRequest struct {
	opsmodels.Request `json:",inline" yaml:",inline"`
}

type PreviewResponse struct {
	Order *opsmodels.ChangeOrder
}

// Preview compute all changes between resources in request and the actual infrastructure.
// The whole process is similar to the operation Apply, but the execution of each node is mocked and will not actually invoke the Runtime
func (po *PreviewOperation) Preview(request *PreviewRequest) (rsp *PreviewResponse, s v1.Status) {
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

	if s := validateRequest(&request.Request); v1.IsErr(s) {
		return nil, s
	}

	var (
		priorState, resultState *states.State
		priorStateResourceIndex map[string]*apiv1.Resource
		ag                      *dag.AcyclicGraph
	)

	// 1. init & build Indexes
	priorState, resultState = po.InitStates(&request.Request)

	// Kusion is a multi-runtime system. We initialize runtimes dynamically by resource types
	resources := request.Intent.Resources
	resources = append(resources, priorState.Resources...)
	runtimesMap, s := runtimeinit.Runtimes(resources)
	if v1.IsErr(s) {
		return nil, s
	}
	o.RuntimeMap = runtimesMap

	switch o.OperationType {
	case opsmodels.ApplyPreview:
		priorStateResourceIndex = priorState.Resources.Index()
		ag, s = NewApplyGraph(request.Intent, priorState)
	case opsmodels.DestroyPreview:
		resources := request.Request.Intent.Resources
		priorStateResourceIndex = resources.Index()
		ag, s = NewDestroyGraph(resources)
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
		Operation: opsmodels.Operation{
			OperationType:           o.OperationType,
			StateStorage:            o.StateStorage,
			CtxResourceIndex:        map[string]*apiv1.Resource{},
			PriorStateResourceIndex: priorStateResourceIndex,
			StateResourceIndex:      stateResourceIndex,
			IgnoreFields:            o.IgnoreFields,
			ChangeOrder:             o.ChangeOrder,
			RuntimeMap:              o.RuntimeMap,
			Stack:                   o.Stack,
			ResultState:             resultState,
			Lock:                    &sync.Mutex{},
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

	if node, ok := v.(graph.ExecutableNode); ok {
		s = node.Execute(&po.Operation)
		if v1.IsErr(s) {
			diags = diags.Append(fmt.Errorf("preview failed.\n%v", s))
			return diags
		}
	}
	return nil
}

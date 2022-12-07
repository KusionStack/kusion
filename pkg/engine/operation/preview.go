package operation

import (
	"errors"
	"fmt"
	"sync"

	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	dag2 "kusionstack.io/kusion/third_party/terraform/dag"
	"kusionstack.io/kusion/third_party/terraform/tfdiags"

	"kusionstack.io/kusion/pkg/engine/operation/graph"
	"kusionstack.io/kusion/pkg/engine/operation/types"

	"kusionstack.io/kusion/pkg/engine/models"

	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/status"
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
func (po *PreviewOperation) Preview(request *PreviewRequest) (rsp *PreviewResponse, s status.Status) {
	o := po.Operation

	defer func() {
		if e := recover(); e != nil {
			log.Error("preview panic:%v", e)

			switch x := e.(type) {
			case string:
				s = status.NewErrorStatus(fmt.Errorf("preview panic:%s", e))
			case error:
				s = status.NewErrorStatus(x)
			default:
				s = status.NewErrorStatus(errors.New("unknown panic"))
			}
		}
	}()

	if s := validateRequest(&request.Request); status.IsErr(s) {
		return nil, s
	}

	var (
		priorState, resultState *states.State
		priorStateResourceIndex map[string]*models.Resource
		ag                      *dag2.AcyclicGraph
	)

	// 1. init & build Indexes
	priorState, resultState = po.InitStates(&request.Request)

	switch o.OperationType {
	case types.ApplyPreview:
		priorStateResourceIndex = priorState.Resources.Index()
		ag, s = NewApplyGraph(request.Spec, priorState)
	case types.DestroyPreview:
		resources := request.Request.Spec.Resources
		priorStateResourceIndex = resources.Index()
		ag, s = NewDestroyGraph(resources)
	}
	if status.IsErr(s) {
		return nil, s
	}

	// 2. walk DAG and preview resources
	log.Info("walking DAG and preview resources ...")

	previewOperation := &PreviewOperation{
		Operation: opsmodels.Operation{
			OperationType:           o.OperationType,
			StateStorage:            o.StateStorage,
			CtxResourceIndex:        map[string]*models.Resource{},
			PriorStateResourceIndex: priorStateResourceIndex,
			StateResourceIndex:      priorStateResourceIndex,
			IgnoreFields:            o.IgnoreFields,
			ChangeOrder:             o.ChangeOrder,
			Runtime:                 o.Runtime, // preview need get the latest spec from runtime
			ResultState:             resultState,
			Lock:                    &sync.Mutex{},
		},
	}

	w := &dag2.Walker{Callback: previewOperation.previewWalkFun}
	w.Update(ag)
	// Wait
	if diags := w.Wait(); diags.HasErrors() {
		return nil, status.NewErrorStatus(diags.Err())
	}

	return &PreviewResponse{Order: previewOperation.ChangeOrder}, nil
}

func (po *PreviewOperation) previewWalkFun(v dag2.Vertex) (diags tfdiags.Diagnostics) {
	var s status.Status
	if v == nil {
		return nil
	}
	defer func() {
		if e := recover(); e != nil {
			log.Errorf("previewWalkFun panic:%v", e)

			var err error
			switch x := e.(type) {
			case string:
				err = fmt.Errorf("previewWalkFun panic:%s", e)
			case error:
				err = x
			default:
				err = errors.New("unknown panic")
			}
			s = status.NewErrorStatus(err)
		}
	}()

	if node, ok := v.(graph.ExecutableNode); ok {
		s = node.Execute(&po.Operation)
		if status.IsErr(s) {
			diags = diags.Append(fmt.Errorf("node execute failed.\n%v", s))
			return diags
		}
	}
	return nil
}

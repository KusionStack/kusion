package operation

import (
	"errors"
	"fmt"
	"sync"

	"kusionstack.io/kusion/pkg/engine/models"

	"github.com/hashicorp/terraform/dag"
	"github.com/hashicorp/terraform/tfdiags"

	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/status"
)

type PreviewOperation struct {
	Operation
}

type PreviewRequest struct {
	Request `json:",inline" yaml:",inline"`
}

type PreviewResponse struct {
	Order *ChangeOrder
}

func (o *Operation) Preview(request *PreviewRequest, operation Type) (rsp *PreviewResponse, s status.Status) {
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
		graph                   *dag.AcyclicGraph
	)

	// 1. init & build Indexes
	priorState, resultState = initStates(o.StateStorage, &request.Request)

	switch operation {
	case Apply:
		priorStateResourceIndex = priorState.Resources.Index()
		graph, s = NewApplyGraph(request.Manifest, priorState)
	case Destroy:
		resources := request.Request.Manifest.Resources
		priorStateResourceIndex = resources.Index()
		graph, s = NewDestroyGraph(resources)
	}
	if status.IsErr(s) {
		return nil, s
	}

	// 2. walk DAG and preview resources
	log.Info("walking DAG and preview resources ...")

	previewOperation := &PreviewOperation{
		Operation: Operation{
			OperationType:           Preview,
			StateStorage:            o.StateStorage,
			CtxResourceIndex:        map[string]*models.Resource{},
			PriorStateResourceIndex: priorStateResourceIndex,
			StateResourceIndex:      priorStateResourceIndex,
			Order:                   o.Order,
			resultState:             resultState,
			lock:                    &sync.Mutex{},
		},
	}

	w := &dag.Walker{Callback: previewOperation.previewWalkFun}
	w.Update(graph)
	// Wait
	if diags := w.Wait(); diags.HasErrors() {
		return nil, status.NewErrorStatus(diags.Err())
	}

	return &PreviewResponse{Order: previewOperation.Order}, nil
}

func (po *PreviewOperation) previewWalkFun(v dag.Vertex) (diags tfdiags.Diagnostics) {
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

	if node, ok := v.(ExecutableNode); ok {
		s = node.Execute(&po.Operation)
		if status.IsErr(s) {
			diags = diags.Append(fmt.Errorf("node execute failed, status: %v", s))
			return diags
		}
	}
	return nil
}

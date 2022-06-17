package operation

import (
	"errors"
	"fmt"
	"sync"

	"kusionstack.io/kusion/pkg/engine/operation/graph"

	"github.com/hashicorp/terraform/tfdiags"

	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"

	"kusionstack.io/kusion/pkg/engine/operation/parser"
	"kusionstack.io/kusion/pkg/engine/operation/types"

	"kusionstack.io/kusion/pkg/engine/models"

	"github.com/hashicorp/terraform/dag"

	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/status"
)

type DestroyOperation struct {
	opsmodels.Operation
}

type DestroyRequest struct {
	opsmodels.Request `json:",inline" yaml:",inline"`
}

func NewDestroyGraph(resource models.Resources) (*dag.AcyclicGraph, status.Status) {
	ag := &dag.AcyclicGraph{}
	ag.Add(&graph.RootNode{})
	deleteResourceParser := parser.NewDeleteResourceParser(resource)
	s := deleteResourceParser.Parse(ag)
	if status.IsErr(s) {
		return nil, s
	}

	return ag, s
}

// Destroy will delete all resources in this request. The whole process is similar to the operation Apply,
// but every node's execution is deleting the resource.
func (do *DestroyOperation) Destroy(request *DestroyRequest) (st status.Status) {
	o := do.Operation

	defer func() {
		close(o.MsgCh)
		if e := recover(); e != nil {
			log.Error("destroy panic:%v", e)

			switch x := e.(type) {
			case string:
				st = status.NewErrorStatus(fmt.Errorf("destroy panic:%s", e))
			case error:
				st = status.NewErrorStatus(x)
			default:
				st = status.NewErrorStatusWithCode(status.Unknown, errors.New("unknown panic"))
			}
		}
	}()

	if st = validateRequest(&request.Request); status.IsErr(st) {
		return st
	}

	// 1. init & build Indexes
	_, resultState := o.InitStates(&request.Request)
	// replace priorState.Resources with models.Resources, so we do Delete in all nodes
	resources := request.Request.Spec.Resources
	priorStateResourceIndex := resources.Index()

	// 2. build & walk DAG
	destroyGraph, s := NewDestroyGraph(resources)
	if status.IsErr(s) {
		return s
	}

	newDo := &DestroyOperation{
		Operation: opsmodels.Operation{
			OperationType:           types.Destroy,
			StateStorage:            o.StateStorage,
			CtxResourceIndex:        map[string]*models.Resource{},
			PriorStateResourceIndex: priorStateResourceIndex,
			StateResourceIndex:      priorStateResourceIndex,
			Runtime:                 o.Runtime,
			MsgCh:                   o.MsgCh,
			ResultState:             resultState,
			Lock:                    &sync.Mutex{},
		},
	}

	w := &dag.Walker{Callback: newDo.destroyWalkFun}
	w.Update(destroyGraph)
	// Wait
	if diags := w.Wait(); diags.HasErrors() {
		st = status.NewErrorStatus(diags.Err())
		return st
	}
	return nil
}

func (do *DestroyOperation) destroyWalkFun(v dag.Vertex) (diags tfdiags.Diagnostics) {
	ao := &ApplyOperation{
		Operation: do.Operation,
	}
	return ao.applyWalkFun(v)
}

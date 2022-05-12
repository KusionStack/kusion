package operation

import (
	"errors"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform/dag"

	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/status"
)

type DestroyOperation struct {
	Operation
}

type DestroyRequest struct {
	Request `json:",inline" yaml:",inline"`
}

func NewDestroyGraph(resource states.Resources) (*dag.AcyclicGraph, status.Status) {
	graph := &dag.AcyclicGraph{}
	graph.Add(&RootNode{})
	deleteResourceParser := NewDeleteResourceParser(resource)
	s := deleteResourceParser.Parse(graph)
	if status.IsErr(s) {
		return nil, s
	}

	return graph, s
}

func (o *Operation) Destroy(request *DestroyRequest) (st status.Status) {
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
	_, resultState := initStates(o.StateStorage, &request.Request)
	// replace priorState.Resources with manifest.Resources, so we do Delete in all nodes
	resources := request.Request.Manifest.Resources
	priorStateResourceIndex := resources.Index()

	// 2. build & walk DAG
	graph, s := NewDestroyGraph(resources)
	if status.IsErr(s) {
		return s
	}

	do := &DestroyOperation{
		Operation: Operation{
			OperationType:           Destroy,
			StateStorage:            o.StateStorage,
			CtxResourceIndex:        map[string]*states.ResourceState{},
			PriorStateResourceIndex: priorStateResourceIndex,
			StateResourceIndex:      priorStateResourceIndex,
			ChangeStepMap:           o.ChangeStepMap,
			Runtime:                 o.Runtime,
			MsgCh:                   o.MsgCh,
			resultState:             resultState,
			lock:                    &sync.Mutex{},
		},
	}

	w := dag.Walker{Callback: do.applyWalkFun}
	w.Update(graph)
	// Wait
	if diags := w.Wait(); diags.HasErrors() {
		st = status.NewErrorStatus(diags.Err())
		return st
	}
	return nil
}

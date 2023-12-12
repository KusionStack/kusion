package operation

import (
	"sync"

	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/status"
	"kusionstack.io/kusion/pkg/engine/operation/graph"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/operation/parser"
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/third_party/terraform/dag"
	"kusionstack.io/kusion/third_party/terraform/tfdiags"
)

type DestroyOperation struct {
	opsmodels.Operation
}

type DestroyRequest struct {
	opsmodels.Request `json:",inline" yaml:",inline"`
}

func NewDestroyGraph(resource intent.Resources) (*dag.AcyclicGraph, status.Status) {
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
	defer close(o.MsgCh)

	if st = validateRequest(&request.Request); status.IsErr(st) {
		return st
	}

	// 1. init & build Indexes
	priorState, resultState := o.InitStates(&request.Request)
	priorStateResourceIndex := priorState.Resources.Index()
	// copy priorStateResourceIndex into a new map
	stateResourceIndex := map[string]*intent.Resource{}
	for k, v := range priorStateResourceIndex {
		stateResourceIndex[k] = v
	}

	// only destroy resources we have recorded
	resources := priorState.Resources
	runtimesMap, s := runtimeinit.Runtimes(resources)
	if status.IsErr(s) {
		return s
	}
	o.RuntimeMap = runtimesMap

	// 2. build & walk DAG
	destroyGraph, s := NewDestroyGraph(resources)
	if status.IsErr(s) {
		return s
	}

	newDo := &DestroyOperation{
		Operation: opsmodels.Operation{
			OperationType:           opsmodels.Destroy,
			StateStorage:            o.StateStorage,
			CtxResourceIndex:        map[string]*intent.Resource{},
			PriorStateResourceIndex: priorStateResourceIndex,
			StateResourceIndex:      stateResourceIndex,
			RuntimeMap:              o.RuntimeMap,
			Stack:                   o.Stack,
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

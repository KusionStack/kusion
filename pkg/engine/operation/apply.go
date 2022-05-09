package operation

import (
	"errors"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform/dag"
	"github.com/hashicorp/terraform/tfdiags"

	"kusionstack.io/kusion/pkg/engine/manifest"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/status"
)

type ApplyOperation struct {
	Operation
}

type ApplyRequest struct {
	Request `json:",inline" yaml:",inline"`
}

type ApplyResponse struct {
	State *states.State
}

func NewApplyGraph(m *manifest.Manifest, priorState *states.State) (*dag.AcyclicGraph, status.Status) {
	manifestParser := NewManifestParser(m)
	graph := &dag.AcyclicGraph{}
	graph.Add(&RootNode{})

	s := manifestParser.Parse(graph)
	if status.IsErr(s) {
		return nil, s
	}
	deleteResourceParser := NewDeleteResourceParser(priorState.Resources)
	s = deleteResourceParser.Parse(graph)
	if status.IsErr(s) {
		return nil, s
	}

	return graph, s
}

func (o *Operation) Apply(request *ApplyRequest) (rsp *ApplyResponse, st status.Status) {
	log.Infof("engine Apply start!")

	defer func() {
		close(o.MsgCh)

		if e := recover(); e != nil {
			log.Error("apply panic:%v", e)

			switch x := e.(type) {
			case string:
				st = status.NewErrorStatus(fmt.Errorf("apply panic:%s", e))
			case error:
				st = status.NewErrorStatus(x)
			default:
				st = status.NewErrorStatusWithCode(status.Unknown, errors.New("unknown panic"))
			}
		}
	}()

	if st = validateRequest(&request.Request); status.IsErr(st) {
		return nil, st
	}

	// 1. init & build Indexes
	priorState, resultState := initStates(o.StateStorage, &request.Request)
	priorStateResourceIndex := priorState.Resources.Index()

	// 2. build & walk DAG
	graph, s := NewApplyGraph(request.Manifest, priorState)
	if status.IsErr(s) {
		return nil, s
	}
	log.Infof("Apply Graph:%s", graph.String())

	applyOperation := &ApplyOperation{
		Operation: Operation{
			OperationType:           Apply,
			StateStorage:            o.StateStorage,
			CtxResourceIndex:        map[string]*states.ResourceState{},
			PriorStateResourceIndex: priorStateResourceIndex,
			StateResourceIndex:      priorStateResourceIndex,
			Runtime:                 o.Runtime,
			MsgCh:                   o.MsgCh,
			resultState:             resultState,
			lock:                    &sync.Mutex{},
		},
	}

	w := &dag.Walker{Callback: applyOperation.applyWalkFun}
	w.Update(graph)
	// Wait
	if diags := w.Wait(); diags.HasErrors() {
		st = status.NewErrorStatus(diags.Err())
		return nil, st
	}

	return &ApplyResponse{State: resultState}, nil
}

func (o *Operation) applyWalkFun(v dag.Vertex) (diags tfdiags.Diagnostics) {
	var s status.Status
	if v == nil {
		return nil
	}

	defer func() {
		if e := recover(); e != nil {
			log.Errorf("applyWalkFun panic:%v", e)

			var err error
			switch x := e.(type) {
			case string:
				err = fmt.Errorf("applyWalkFun panic:%s", e)
			case error:
				err = x
			default:
				err = errors.New("unknown panic")
			}
			s = status.NewErrorStatus(err)
		}
	}()

	if node, ok := v.(ExecutableNode); ok {
		if rn, ok2 := v.(*ResourceNode); ok2 {
			o.MsgCh <- Message{rn.Hashcode().(string), "", nil}

			s = node.Execute(o)
			if status.IsErr(s) {
				o.MsgCh <- Message{rn.Hashcode().(string), Failed, fmt.Errorf("node execte failed, status: %v", s)}
			} else {
				o.MsgCh <- Message{rn.Hashcode().(string), Success, nil}
			}
		} else {
			s = node.Execute(o)
		}
	}
	if s != nil {
		diags = diags.Append(fmt.Errorf("node execte failed, status: %v", s))
	}
	return diags
}

func validateRequest(request *Request) status.Status {
	var s status.Status

	if request == nil {
		return status.NewErrorStatusWithMsg(status.InvalidArgument, "request is nil")
	}
	if request.Manifest == nil {
		return status.NewErrorStatusWithMsg(status.InvalidArgument,
			"request.Manifest is empty. If you want to delete all resources, please use command 'destroy'")
	}
	resourceKeyMap := make(map[string]bool)

	for _, resource := range request.Manifest.Resources {
		key := resource.ResourceKey()
		if _, ok := resourceKeyMap[key]; ok {
			return status.NewErrorStatusWithMsg(status.InvalidArgument, fmt.Sprintf("Duplicate resource:%s in request.", key))
		}
		resourceKeyMap[key] = true
	}

	return s
}

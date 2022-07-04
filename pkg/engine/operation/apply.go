package operation

import (
	"errors"
	"fmt"
	"sync"

	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"

	"kusionstack.io/kusion/pkg/engine/operation/graph"
	"kusionstack.io/kusion/pkg/engine/operation/parser"
	"kusionstack.io/kusion/pkg/engine/operation/types"

	"github.com/hashicorp/terraform/dag"
	"github.com/hashicorp/terraform/tfdiags"

	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/status"
)

type ApplyOperation struct {
	opsmodels.Operation
}

type ApplyRequest struct {
	opsmodels.Request `json:",inline" yaml:",inline"`
}

type ApplyResponse struct {
	State *states.State
}

func NewApplyGraph(m *models.Spec, priorState *states.State) (*dag.AcyclicGraph, status.Status) {
	specParser := parser.NewSpecParser(m)
	g := &dag.AcyclicGraph{}
	g.Add(&graph.RootNode{})

	s := specParser.Parse(g)
	if status.IsErr(s) {
		return nil, s
	}
	deleteResourceParser := parser.NewDeleteResourceParser(priorState.Resources)
	s = deleteResourceParser.Parse(g)
	if status.IsErr(s) {
		return nil, s
	}

	return g, s
}

// Apply means turn all actual infra resources into the desired state described in the request by invoking a specified Runtime.
// Like other operations, Apply has 3 main steps during the whole process.
//  1. parse resources and their relationship to build a DAG and should take care of those resources that will be deleted
//  2. walk this DAG and execute all graph nodes concurrently, besides the entire process should follow dependencies in this DAG
//  3. during the execution of each node, it will invoke different runtime according to the resource type
func (ao *ApplyOperation) Apply(request *ApplyRequest) (rsp *ApplyResponse, st status.Status) {
	log.Infof("engine: Apply start!")
	o := ao.Operation

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
	priorState, resultState := o.InitStates(&request.Request)
	priorStateResourceIndex := priorState.Resources.Index()

	// 2. build & walk DAG
	applyGraph, s := NewApplyGraph(request.Spec, priorState)
	if status.IsErr(s) {
		return nil, s
	}
	log.Infof("Apply Graph:\n%s", applyGraph.String())

	applyOperation := &ApplyOperation{
		Operation: opsmodels.Operation{
			OperationType:           types.Apply,
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

	w := &dag.Walker{Callback: applyOperation.applyWalkFun}
	w.Update(applyGraph)
	// Wait
	if diags := w.Wait(); diags.HasErrors() {
		st = status.NewErrorStatus(diags.Err())
		return nil, st
	}

	return &ApplyResponse{State: resultState}, nil
}

func (ao *ApplyOperation) applyWalkFun(v dag.Vertex) (diags tfdiags.Diagnostics) {
	var s status.Status
	if v == nil {
		return nil
	}
	o := &ao.Operation

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

	if node, ok := v.(graph.ExecutableNode); ok {
		if rn, ok2 := v.(*graph.ResourceNode); ok2 {
			o.MsgCh <- opsmodels.Message{ResourceID: rn.Hashcode().(string)}

			s = node.Execute(o)
			if status.IsErr(s) {
				o.MsgCh <- opsmodels.Message{ResourceID: rn.Hashcode().(string), OpResult: opsmodels.Failed, OpErr: fmt.Errorf("node execte failed, status: %v", s)}
			} else {
				o.MsgCh <- opsmodels.Message{ResourceID: rn.Hashcode().(string), OpResult: opsmodels.Success}
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

func validateRequest(request *opsmodels.Request) status.Status {
	var s status.Status

	if request == nil {
		return status.NewErrorStatusWithMsg(status.InvalidArgument, "request is nil")
	}
	if request.Spec == nil {
		return status.NewErrorStatusWithMsg(status.InvalidArgument,
			"request.Spec is empty. If you want to delete all resources, please use command 'destroy'")
	}
	resourceKeyMap := make(map[string]bool)

	for _, resource := range request.Spec.Resources {
		key := resource.ResourceKey()
		if _, ok := resourceKeyMap[key]; ok {
			return status.NewErrorStatusWithMsg(status.InvalidArgument, fmt.Sprintf("Duplicate resource:%s in request.", key))
		}
		resourceKeyMap[key] = true
	}

	return s
}

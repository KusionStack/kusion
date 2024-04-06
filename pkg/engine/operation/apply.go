package operation

import (
	"errors"
	"fmt"
	"sync"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation/graph"
	models "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/operation/parser"
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/third_party/terraform/dag"
	"kusionstack.io/kusion/third_party/terraform/tfdiags"
)

type ApplyOperation struct {
	models.Operation
}

type ApplyRequest struct {
	models.Request `json:",inline" yaml:",inline"`
}

type ApplyResponse struct {
	State *apiv1.State
}

func NewApplyGraph(m *apiv1.Spec, priorState *apiv1.State) (*dag.AcyclicGraph, v1.Status) {
	intentParser := parser.NewIntentParser(m)
	g := &dag.AcyclicGraph{}
	g.Add(&graph.RootNode{})

	s := intentParser.Parse(g)
	if v1.IsErr(s) {
		return nil, s
	}
	deleteResourceParser := parser.NewDeleteResourceParser(priorState.Resources)
	s = deleteResourceParser.Parse(g)
	if v1.IsErr(s) {
		return nil, s
	}

	return g, s
}

// Apply means turn all actual infra resources into the desired state described in the request by invoking a specified Runtime.
// Like other operations, Apply has 3 main steps during the whole process.
//  1. parse resources and their relationship to build a DAG and should take care of those resources that will be deleted
//  2. walk this DAG and execute all graph nodes concurrently, besides the entire process should follow dependencies in this DAG
//  3. during the execution of each node, it will invoke different runtime according to the resource type
func (ao *ApplyOperation) Apply(request *ApplyRequest) (rsp *ApplyResponse, st v1.Status) {
	log.Infof("engine: Apply start!")
	o := ao.Operation

	defer func() {
		close(o.MsgCh)

		if e := recover(); e != nil {
			log.Error("apply panic:%v", e)

			switch x := e.(type) {
			case string:
				st = v1.NewErrorStatus(fmt.Errorf("apply panic:%s", e))
			case error:
				st = v1.NewErrorStatus(x)
			default:
				st = v1.NewErrorStatusWithCode(v1.Unknown, errors.New("unknown panic"))
			}
		}
	}()

	if st = validateRequest(&request.Request); v1.IsErr(st) {
		return nil, st
	}

	// 1. init & build Indexes
	priorState, resultState := o.InitStates(&request.Request)
	priorStateResourceIndex := priorState.Resources.Index()
	// copy priorStateResourceIndex into a new map
	stateResourceIndex := map[string]*apiv1.Resource{}
	for k, v := range priorStateResourceIndex {
		stateResourceIndex[k] = v
	}

	resources := request.Intent.Resources
	resources = append(resources, priorState.Resources...)
	runtimesMap, s := runtimeinit.Runtimes(resources)
	if v1.IsErr(s) {
		return nil, s
	}
	o.RuntimeMap = runtimesMap

	// 2. build & walk DAG
	applyGraph, s := NewApplyGraph(request.Intent, priorState)
	if v1.IsErr(s) {
		return nil, s
	}
	log.Infof("Apply Graph:\n%s", applyGraph.String())

	applyOperation := &ApplyOperation{
		Operation: models.Operation{
			OperationType:           models.Apply,
			StateStorage:            o.StateStorage,
			CtxResourceIndex:        map[string]*apiv1.Resource{},
			PriorStateResourceIndex: priorStateResourceIndex,
			StateResourceIndex:      stateResourceIndex,
			RuntimeMap:              o.RuntimeMap,
			Stack:                   o.Stack,
			IgnoreFields:            o.IgnoreFields,
			MsgCh:                   o.MsgCh,
			ResultState:             resultState,
			Lock:                    &sync.Mutex{},
		},
	}

	w := &dag.Walker{Callback: applyOperation.applyWalkFun}
	w.Update(applyGraph)
	// Wait
	if diags := w.Wait(); diags.HasErrors() {
		st = v1.NewErrorStatus(diags.Err())
		return nil, st
	}

	return &ApplyResponse{State: resultState}, nil
}

func (ao *ApplyOperation) applyWalkFun(v dag.Vertex) (diags tfdiags.Diagnostics) {
	var s v1.Status
	if v == nil {
		return nil
	}
	o := &ao.Operation

	if node, ok := v.(graph.ExecutableNode); ok {
		if rn, ok2 := v.(*graph.ResourceNode); ok2 {
			o.MsgCh <- models.Message{ResourceID: rn.Hashcode().(string)}

			s = node.Execute(o)
			if v1.IsErr(s) {
				o.MsgCh <- models.Message{
					ResourceID: rn.Hashcode().(string), OpResult: models.Failed,
					OpErr: fmt.Errorf("node execte failed, status:\n%v", s),
				}
			} else {
				o.MsgCh <- models.Message{ResourceID: rn.Hashcode().(string), OpResult: models.Success}
			}
		} else {
			s = node.Execute(o)
		}
	}
	if s != nil {
		diags = diags.Append(fmt.Errorf("apply failed, status:\n%v", s))
	}
	return diags
}

func validateRequest(request *models.Request) v1.Status {
	var s v1.Status

	if request == nil {
		return v1.NewErrorStatusWithMsg(v1.InvalidArgument, "request is nil")
	}
	if request.Intent == nil {
		return v1.NewErrorStatusWithMsg(v1.InvalidArgument,
			"request.Intent is empty. If you want to delete all resources, please use command 'destroy'")
	}
	resourceKeyMap := make(map[string]bool)

	for _, resource := range request.Intent.Resources {
		key := resource.ResourceKey()
		if _, ok := resourceKeyMap[key]; ok {
			return v1.NewErrorStatusWithMsg(v1.InvalidArgument, fmt.Sprintf("Duplicate resource:%s in request.", key))
		}
		resourceKeyMap[key] = true
	}

	return s
}

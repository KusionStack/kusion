package operation

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/status"
	jsonUtil "kusionstack.io/kusion/pkg/util/json"
)

type ResourceNode struct {
	BaseNode
	Action ActionType
	state  *states.ResourceState
}

var _ ExecutableNode = (*ResourceNode)(nil)

func (rn *ResourceNode) Execute(operation Operation) status.Status {
	log.Debugf("execute node:%s", rn.ID)
	// 1. prepare planedState
	planedState := rn.state
	if rn.Action != Delete {
		if operation.OperationType != Preview {
			// replace implicit references
			value := reflect.ValueOf(rn.state.Attributes)
			_, implicitValue, s := ParseImplicitRef(value, operation.CtxResourceIndex, ImplicitReplaceFun)
			if status.IsErr(s) {
				return s
			}
			rn.state.Attributes = implicitValue.Interface().(map[string]interface{})
		}
	} else {
		planedState = nil
	}

	// 2. get State
	key := rn.state.ResourceKey()
	priorState := operation.PriorStateResourceIndex[key]

	// 3. compute ActionType of current resource node
	if priorState == nil {
		rn.Action = Create
	} else if planedState == nil {
		rn.Action = Delete
	} else if reflect.DeepEqual(priorState, planedState) {
		rn.Action = UnChange
	} else {
		rn.Action = Update
	}
	fillResponseChangeSteps(operation, rn, priorState, planedState)

	// 4. apply
	if operation.OperationType == Preview {
		log.Infof("operation type is Preview. Skip this resource")
		return nil
	}
	switch rn.Action {
	case Create, Delete, Update:
		s := rn.applyResource(operation, priorState, planedState)
		if status.IsErr(s) {
			return s
		}
	case UnChange:
		log.Infof("PriorAttributes and PlanAttributes are equal.")
	default:
		return status.NewErrorStatus(fmt.Errorf("unknown op:%s", rn.Action.PrettyString()))
	}
	return nil
}

func (rn *ResourceNode) applyResource(operation Operation, priorState *states.ResourceState, planedState *states.ResourceState) status.Status {
	log.Infof("PriorAttributes and PlanAttributes are not equal. operation:%v, prior:%v, plan:%v", rn.Action,
		jsonUtil.Marshal2String(priorState), jsonUtil.Marshal2String(planedState))
	var res *states.ResourceState
	var s status.Status
	switch rn.Action {
	case Create, Update:
		res, s = operation.Runtime.Apply(context.Background(), priorState, planedState)
		log.Debugf("apply resource:%s, result: %v", planedState.ID, jsonUtil.Marshal2String(res))
		if s != nil {
			log.Debugf("apply status: %v", s.String())
		}
	case Delete:
		s = operation.Runtime.Delete(context.Background(), priorState)
		if s != nil {
			log.Debugf("delete state: %v", s.String())
		}
	}
	if status.IsErr(s) {
		return s
	}

	// compatible with delete action
	if res != nil {
		res.DependsOn = planedState.DependsOn
	}
	key := rn.state.ResourceKey()
	if e := operation.RefreshResourceIndex(key, res, rn.Action); e != nil {
		return status.NewErrorStatus(e)
	}
	if e := operation.UpdateState(operation.StateResourceIndex); e != nil {
		return status.NewErrorStatus(e)
	}

	// print apply resource success msg
	log.Infof("apply resource success: %s", rn.state.ResourceKey())
	return nil
}

func (rn *ResourceNode) State() *states.ResourceState {
	return rn.state
}

func NewResourceNode(key string, state *states.ResourceState, action ActionType) *ResourceNode {
	return &ResourceNode{BaseNode: BaseNode{ID: key}, Action: action, state: state}
}

func fillResponseChangeSteps(operation Operation, rn *ResourceNode, prior, plan interface{}) {
	defer operation.lock.Unlock()
	operation.lock.Lock()

	if operation.ChangeStepMap == nil {
		operation.ChangeStepMap = make(map[string]*ChangeStep)
	}
	operation.ChangeStepMap[rn.ID] = NewChangeStep(rn.ID, rn.Action, prior, plan)
}

var ImplicitReplaceFun = func(resourceIndex map[string]*states.ResourceState, refPath string) (reflect.Value, status.Status) {
	const Sep = "."
	split := strings.Split(refPath, Sep)
	key := split[0]
	priorState := resourceIndex[key]
	if priorState == nil {
		msg := fmt.Sprintf("can't find state by key:%s when replacing %s", key, refPath)
		return reflect.Value{}, status.NewErrorStatusWithMsg(status.IllegalManifest, msg)
	}
	attributes := priorState.Attributes
	if attributes == nil {
		msg := fmt.Sprintf("attributes is nil in resource:%s", key)
		return reflect.Value{}, status.NewErrorStatusWithMsg(status.IllegalManifest, msg)
	}
	var valueMap interface{}
	valueMap = attributes
	if len(split) > 1 {
		split := split[1:]
		for _, k := range split {
			if valueMap.(map[string]interface{})[k] == nil {
				msg := fmt.Sprintf("can't find specified value in resource:%s by ref:%s", key, refPath)
				return reflect.Value{}, status.NewErrorStatusWithMsg(status.IllegalManifest, msg)
			}
			valueMap = valueMap.(map[string]interface{})[k]
		}
	}
	return reflect.ValueOf(valueMap), nil
}

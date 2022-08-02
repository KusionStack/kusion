package graph

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"kusionstack.io/kusion/pkg/engine/models"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/operation/types"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/status"
	"kusionstack.io/kusion/pkg/util"
	"kusionstack.io/kusion/pkg/util/diff"
	jsonutil "kusionstack.io/kusion/pkg/util/json"
)

type ResourceNode struct {
	*baseNode
	Action types.ActionType
	state  *models.Resource
}

var _ ExecutableNode = (*ResourceNode)(nil)

const (
	ImplicitRefPrefix = "$kusion_path."
)

func (rn *ResourceNode) Execute(operation *opsmodels.Operation) status.Status {
	log.Debugf("execute node:%s", rn.ID)
	if operation.OperationType == types.Apply {
		// replace implicit references
		value := reflect.ValueOf(rn.state.Attributes)
		_, implicitValue, s := ParseImplicitRef(value, operation.CtxResourceIndex, ImplicitReplaceFun)
		if status.IsErr(s) {
			return s
		}
		rn.state.Attributes = implicitValue.Interface().(map[string]interface{})
	}

	// 1. prepare planedState
	planedState := rn.state
	if rn.Action == types.Delete {
		planedState = nil
	}
	// predictableState represents dry-run result
	predictableState := planedState

	// 2. get prior state which is stored in kusion_state.json
	key := rn.state.ResourceKey()
	priorState := operation.PriorStateResourceIndex[key]

	// 3. get the latest resource from runtime
	readRequest := &runtime.ReadRequest{Resource: planedState}
	if readRequest.Resource == nil {
		readRequest.Resource = priorState
	}
	response := operation.Runtime.Read(context.Background(), readRequest)
	liveState := response.Resource
	s := response.Status
	if status.IsErr(s) {
		return s
	}

	// 4. compute ActionType of current resource node between planState and liveState
	switch operation.OperationType {
	case types.Destroy, types.DestroyPreview:
		rn.Action = types.Delete
	case types.Apply, types.ApplyPreview:
		if liveState == nil {
			rn.Action = types.Create
		} else if planedState == nil {
			rn.Action = types.Delete
		} else {
			// Dry run to fetch predictable state
			dryRunResp := operation.Runtime.Apply(context.Background(), &runtime.ApplyRequest{
				PriorResource: priorState,
				PlanResource:  planedState,
				DryRun:        true,
			})
			if status.IsErr(dryRunResp.Status) {
				return dryRunResp.Status
			}
			predictableState = dryRunResp.Resource
			report, err := diff.ToReport(liveState, predictableState)
			if err != nil {
				return status.NewErrorStatus(err)
			}
			if len(report.Diffs) == 0 {
				rn.Action = types.UnChange
			} else {
				rn.Action = types.Update
			}
		}
	default:
		return status.NewErrorStatus(fmt.Errorf("unknown operation: %v", operation.OperationType))
	}

	// 5. apply or return
	switch operation.OperationType {
	case types.ApplyPreview, types.DestroyPreview:
		fillResponseChangeSteps(operation, rn, liveState, predictableState)
	case types.Apply, types.Destroy:
		if s = rn.applyResource(operation, priorState, planedState); status.IsErr(s) {
			return s
		}
	default:
		return status.NewErrorStatus(fmt.Errorf("unknown operation: %v", operation.OperationType))
	}

	return nil
}

func (rn *ResourceNode) applyResource(operation *opsmodels.Operation, priorState, planedState *models.Resource) status.Status {
	log.Infof("operation:%v, prior:%v, plan:%v, live:%v", rn.Action, jsonutil.Marshal2String(priorState),
		jsonutil.Marshal2String(planedState))

	var res *models.Resource
	var s status.Status

	switch rn.Action {
	case types.Create, types.Update:
		response := operation.Runtime.Apply(context.Background(), &runtime.ApplyRequest{PriorResource: priorState, PlanResource: planedState})
		res = response.Resource
		s = response.Status
		log.Debugf("apply resource:%s, result: %v", planedState.ID, jsonutil.Marshal2String(res))
		if s != nil {
			log.Debugf("apply status: %v", s.String())
		}
	case types.Delete:
		response := operation.Runtime.Delete(context.Background(), &runtime.DeleteRequest{Resource: priorState})
		s = response.Status
		if s != nil {
			log.Debugf("delete state: %v", s.String())
		}
	case types.UnChange:
		log.Infof("planed resource not update live state")
		res = planedState
	}
	if status.IsErr(s) {
		return s
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

func (rn *ResourceNode) State() *models.Resource {
	return rn.state
}

func NewResourceNode(key string, state *models.Resource, action types.ActionType) (*ResourceNode, status.Status) {
	node, s := NewBaseNode(key)
	if status.IsErr(s) {
		return nil, s
	}
	return &ResourceNode{baseNode: node, Action: action, state: state}, nil
}

// save change steps in DAG walking order so that we can preview a full applying list
func fillResponseChangeSteps(ops *opsmodels.Operation, rn *ResourceNode, plan, live interface{}) {
	defer ops.Lock.Unlock()
	ops.Lock.Lock()

	order := ops.ChangeOrder
	if order == nil {
		order = &opsmodels.ChangeOrder{
			StepKeys:    []string{},
			ChangeSteps: make(map[string]*opsmodels.ChangeStep),
		}
	}
	if order.ChangeSteps == nil {
		order.ChangeSteps = make(map[string]*opsmodels.ChangeStep)
	}
	order.StepKeys = append(order.StepKeys, rn.ID)
	order.ChangeSteps[rn.ID] = opsmodels.NewChangeStep(rn.ID, rn.Action, plan, live)
}

var ImplicitReplaceFun = func(resourceIndex map[string]*models.Resource, refPath string) (reflect.Value, status.Status) {
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

func ParseImplicitRef(v reflect.Value, resourceIndex map[string]*models.Resource,
	replaceFun func(resourceIndex map[string]*models.Resource, refPath string) (reflect.Value, status.Status),
) ([]string, reflect.Value, status.Status) {
	var result []string
	if !v.IsValid() {
		return nil, v, status.NewErrorStatusWithMsg(status.InvalidArgument, "invalid implicit reference")
	}

	switch v.Type().Kind() {
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return nil, v, nil
		}
		return ParseImplicitRef(v.Elem(), resourceIndex, replaceFun)
	case reflect.String:
		vStr := v.String()
		if strings.HasPrefix(vStr, ImplicitRefPrefix) {
			ref := strings.TrimPrefix(vStr, ImplicitRefPrefix)
			util.CheckArgument(len(ref) > 0,
				fmt.Sprintf("illegal implicit ref:%s. Implicit ref format: %sresourceKey.attribute", ref, ImplicitRefPrefix))
			split := strings.Split(ref, ".")
			result = append(result, split[0])
			log.Infof("add implicit ref:%s", split[0])
			// replace v with output
			tv, s := replaceFun(resourceIndex, ref)
			if status.IsErr(s) {
				return nil, v, s
			}
			v = tv
		}
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return nil, v, nil
		}

		vs := reflect.MakeSlice(v.Type(), 0, 0)

		for i := 0; i < v.Len(); i++ {
			ref, tv, s := ParseImplicitRef(v.Index(i), resourceIndex, replaceFun)
			if status.IsErr(s) {
				return nil, tv, s
			}
			vs = reflect.Append(vs, tv)
			if ref != nil {
				result = append(result, ref...)
			}
		}
		v = vs
	case reflect.Map:
		if v.Len() == 0 {
			return nil, v, nil
		}
		makeMap := reflect.MakeMap(v.Type())

		iter := v.MapRange()
		for iter.Next() {
			ref, tv, s := ParseImplicitRef(iter.Value(), resourceIndex, replaceFun)
			if status.IsErr(s) {
				return nil, tv, s
			}
			if ref != nil {
				result = append(result, ref...)
			}
			makeMap.SetMapIndex(iter.Key(), tv)
		}
		v = makeMap
	}
	return result, v, nil
}

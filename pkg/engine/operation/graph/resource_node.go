package graph

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"kusionstack.io/kusion/pkg/engine/models"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/status"
	"kusionstack.io/kusion/pkg/util"
	"kusionstack.io/kusion/pkg/util/diff"
	jsonutil "kusionstack.io/kusion/pkg/util/json"
	"kusionstack.io/kusion/pkg/vals"
)

type ResourceNode struct {
	*baseNode
	Action opsmodels.ActionType
	state  *models.Resource
}

var _ ExecutableNode = (*ResourceNode)(nil)

const (
	ImplicitRefPrefix = "$kusion_path."
)

func (rn *ResourceNode) PreExecute(o *opsmodels.Operation) status.Status {
	value := reflect.ValueOf(rn.state.Attributes)
	var replaced reflect.Value
	var s status.Status
	switch o.OperationType {
	case opsmodels.ApplyPreview:
		// replace secret ref
		_, replaced, s = ReplaceSecretRef(value, o.SecretStores)
	case opsmodels.Apply:
		// replace secret ref and implicit ref
		_, replaced, s = ReplaceRef(value, o.CtxResourceIndex, ImplicitReplaceFun, o.SecretStores, vals.ParseSecretRef)
	default:
		return nil
	}
	if status.IsErr(s) {
		return s
	}
	if !replaced.IsZero() {
		rn.state.Attributes = replaced.Interface().(map[string]interface{})
	}
	return nil
}

func (rn *ResourceNode) Execute(operation *opsmodels.Operation) status.Status {
	log.Debugf("execute node:%s", rn.ID)

	if s := rn.PreExecute(operation); status.IsErr(s) {
		return s
	}

	// 1. prepare planedState
	planedState := rn.state
	if rn.Action == opsmodels.Delete {
		planedState = nil
	}
	// predictableState represents dry-run result
	predictableState := planedState

	// 2. get prior state which is stored in kusion_state.json
	key := rn.state.ResourceKey()
	priorState := operation.PriorStateResourceIndex[key]

	// 3. get the latest resource from runtime
	readRequest := &runtime.ReadRequest{PlanResource: planedState, PriorResource: priorState, Stack: operation.Stack}

	response := operation.Runtime.Read(context.Background(), readRequest)
	liveState := response.Resource
	s := response.Status
	if status.IsErr(s) {
		return s
	}

	// 4. compute ActionType of current resource node between planState and liveState
	switch operation.OperationType {
	case opsmodels.Destroy, opsmodels.DestroyPreview:
		rn.Action = opsmodels.Delete
	case opsmodels.Apply, opsmodels.ApplyPreview:
		if planedState == nil {
			rn.Action = opsmodels.Delete
		} else if priorState == nil && liveState == nil {
			rn.Action = opsmodels.Create
		} else {
			// Dry run to fetch predictable state
			dryRunResp := operation.Runtime.Apply(context.Background(), &runtime.ApplyRequest{
				PriorResource: priorState,
				PlanResource:  planedState,
				Stack:         operation.Stack,
				DryRun:        true,
			})
			if status.IsErr(dryRunResp.Status) {
				return dryRunResp.Status
			}
			predictableState = dryRunResp.Resource
			// Ignore differences of target fields
			for _, field := range operation.IgnoreFields {
				splits := strings.Split(field, ".")
				unstructured.RemoveNestedField(liveState.Attributes, splits...)
				unstructured.RemoveNestedField(predictableState.Attributes, splits...)
			}
			report, err := diff.ToReport(liveState, predictableState)
			if err != nil {
				return status.NewErrorStatus(err)
			}
			if len(report.Diffs) == 0 {
				rn.Action = opsmodels.UnChange
			} else {
				rn.Action = opsmodels.Update
			}
		}
	default:
		return status.NewErrorStatus(fmt.Errorf("unknown operation: %v", operation.OperationType))
	}

	// 5. apply or return
	switch operation.OperationType {
	case opsmodels.ApplyPreview, opsmodels.DestroyPreview:
		fillResponseChangeSteps(operation, rn, liveState, predictableState)
	case opsmodels.Apply, opsmodels.Destroy:
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
	case opsmodels.Create, opsmodels.Update:
		response := operation.Runtime.Apply(context.Background(), &runtime.ApplyRequest{
			PriorResource: priorState, PlanResource: planedState, Stack: operation.Stack,
		})
		res = response.Resource
		s = response.Status
		log.Debugf("apply resource:%s, result: %v", planedState.ID, jsonutil.Marshal2String(res))
		if s != nil {
			log.Debugf("apply status: %v", s.String())
		}
	case opsmodels.Delete:
		response := operation.Runtime.Delete(context.Background(), &runtime.DeleteRequest{Resource: priorState, Stack: operation.Stack})
		s = response.Status
		if s != nil {
			log.Debugf("delete state: %v", s.String())
		}
	case opsmodels.UnChange:
		log.Infof("planed resource not update live state")
		res = priorState
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

func NewResourceNode(key string, state *models.Resource, action opsmodels.ActionType) (*ResourceNode, status.Status) {
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

func ReplaceSecretRef(v reflect.Value, ss *vals.SecretStores) ([]string, reflect.Value, status.Status) {
	return ReplaceRef(v, nil, nil, ss, vals.ParseSecretRef)
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

func ReplaceImplicitRef(v reflect.Value, resourceIndex map[string]*models.Resource,
	replaceFun func(map[string]*models.Resource, string) (reflect.Value, status.Status),
) ([]string, reflect.Value, status.Status) {
	return ReplaceRef(v, resourceIndex, replaceFun, nil, nil)
}

func ReplaceRef(v reflect.Value,
	resourceIndex map[string]*models.Resource,
	replaceImplicitFun func(map[string]*models.Resource, string) (reflect.Value, status.Status),
	ss *vals.SecretStores,
	replaceSecretFun func(string, string, *vals.SecretStores) (string, error),
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
		return ReplaceRef(v.Elem(), resourceIndex, replaceImplicitFun, ss, replaceSecretFun)
	case reflect.String:
		vStr := v.String()
		if replaceImplicitFun != nil {
			if strings.HasPrefix(vStr, ImplicitRefPrefix) {
				ref := strings.TrimPrefix(vStr, ImplicitRefPrefix)
				util.CheckArgument(len(ref) > 0,
					fmt.Sprintf("illegal implicit ref:%s. Implicit ref format: %sresourceKey.attribute", ref, ImplicitRefPrefix))
				split := strings.Split(ref, ".")
				result = append(result, split[0])
				log.Infof("add implicit ref:%s", split[0])
				// replace v with output
				tv, s := replaceImplicitFun(resourceIndex, ref)
				if status.IsErr(s) {
					return nil, v, s
				}
				v = tv
			}
		}

		if ss != nil && replaceSecretFun != nil {
			if prefix, ok := vals.IsSecured(vStr); ok {
				tStr, err := replaceSecretFun(prefix, vStr, ss)
				if err != nil {
					return nil, v, status.NewErrorStatus(err)
				}
				tv := reflect.New(v.Type()).Elem()
				tv.SetString(tStr)
				v = tv
			}
		}
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return nil, v, nil
		}

		vs := reflect.MakeSlice(v.Type(), 0, 0)

		for i := 0; i < v.Len(); i++ {
			ref, tv, s := ReplaceRef(v.Index(i), resourceIndex, replaceImplicitFun, ss, replaceSecretFun)
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
			ref, tv, s := ReplaceRef(iter.Value(), resourceIndex, replaceImplicitFun, ss, replaceSecretFun)
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

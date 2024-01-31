package graph

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util"
	"kusionstack.io/kusion/pkg/util/diff"
	jsonutil "kusionstack.io/kusion/pkg/util/json"
)

type ResourceNode struct {
	*baseNode
	Action   opsmodels.ActionType
	resource *apiv1.Resource
}

var _ ExecutableNode = (*ResourceNode)(nil)

const (
	ImplicitRefPrefix = "$kusion_path."
)

func (rn *ResourceNode) PreExecute(o *opsmodels.Operation) v1.Status {
	value := reflect.ValueOf(rn.resource.Attributes)
	var replaced reflect.Value
	var s v1.Status

	switch o.OperationType {
	case opsmodels.ApplyPreview:
		// first time apply. Do not replace implicit dependency ref
		if len(o.PriorStateResourceIndex) == 0 {
			_, replaced, s = ReplaceSecretRef(value)
		} else {
			_, replaced, s = ReplaceRef(value, o.CtxResourceIndex, OptionalImplicitReplaceFun)
		}
	case opsmodels.Apply:
		// replace secret ref and implicit ref
		_, replaced, s = ReplaceRef(value, o.CtxResourceIndex, MustImplicitReplaceFun)
	default:
		return nil
	}
	if v1.IsErr(s) {
		return s
	}
	if !replaced.IsZero() {
		rn.resource.Attributes = replaced.Interface().(map[string]interface{})
	}
	return nil
}

func (rn *ResourceNode) Execute(operation *opsmodels.Operation) (s v1.Status) {
	log.Debugf("executing resource node:%s", rn.ID)

	defer func() {
		log.Debugf("resource node:%s has been executed", rn.ID)

		if e := recover(); e != nil {
			log.Errorf("resource node execution panic:%v", e)

			var err error
			switch x := e.(type) {
			case string:
				err = fmt.Errorf("resource node execution panic:%s", e)
			case error:
				err = x
			default:
				err = errors.New("unknown panic")
			}
			s = v1.NewErrorStatus(err)
		}
	}()

	if s = rn.PreExecute(operation); v1.IsErr(s) {
		return s
	}

	// init 3-way diff data
	planedResource, priorResource, liveResource, s := rn.initThreeWayDiffData(operation)
	if v1.IsErr(s) {
		return s
	}

	// compute action type
	dryRunResource, s := rn.computeActionType(operation, planedResource, priorResource, liveResource)
	if v1.IsErr(s) {
		return s
	}

	// execute the operation
	switch operation.OperationType {
	case opsmodels.ApplyPreview, opsmodels.DestroyPreview:
		key := rn.resource.ResourceKey()
		// refresh resource index in operation to make sure other resource node can get the latest index
		if e := operation.RefreshResourceIndex(key, dryRunResource, rn.Action); e != nil {
			return v1.NewErrorStatus(e)
		}
		updateChangeOrder(operation, rn, liveResource, dryRunResource)
	case opsmodels.Apply, opsmodels.Destroy:
		if s = rn.applyResource(operation, priorResource, planedResource, liveResource); v1.IsErr(s) {
			return s
		}
	default:
		return v1.NewErrorStatus(fmt.Errorf("unknown operation: %v", operation.OperationType))
	}

	return nil
}

// computeActionType compute ActionType of current resource node according to  planResource, priorResource and liveResource.
// dryRunResource is a middle result during the process of computing ActionType. We will use it to perform live diff latter
func (rn *ResourceNode) computeActionType(
	operation *opsmodels.Operation,
	planedResource *apiv1.Resource,
	priorResource *apiv1.Resource,
	liveResource *apiv1.Resource,
) (*apiv1.Resource, v1.Status) {
	dryRunResource := planedResource
	switch operation.OperationType {
	case opsmodels.Destroy, opsmodels.DestroyPreview:
		rn.Action = opsmodels.Delete
	case opsmodels.Apply, opsmodels.ApplyPreview:
		if planedResource == nil {
			rn.Action = opsmodels.Delete
		} else if liveResource == nil {
			rn.Action = opsmodels.Create
		} else {
			// Dry run to fetch predictable resource
			dryRunResp := operation.RuntimeMap[rn.resource.Type].Apply(context.Background(), &runtime.ApplyRequest{
				PriorResource: priorResource,
				PlanResource:  planedResource,
				Stack:         operation.Stack,
				DryRun:        true,
			})
			if v1.IsErr(dryRunResp.Status) {
				return nil, dryRunResp.Status
			}
			dryRunResource = dryRunResp.Resource
			// Ignore differences of target fields
			for _, field := range operation.IgnoreFields {
				splits := strings.Split(field, ".")
				removeNestedField(liveResource.Attributes, splits...)
				removeNestedField(dryRunResource.Attributes, splits...)
			}
			report, err := diff.ToReport(liveResource, dryRunResource)
			if err != nil {
				return nil, v1.NewErrorStatus(err)
			}
			if len(report.Diffs) == 0 {
				rn.Action = opsmodels.UnChanged
			} else {
				rn.Action = opsmodels.Update
			}
		}
	default:
		return nil, v1.NewErrorStatus(fmt.Errorf("unknown operation: %v", operation.OperationType))
	}
	return dryRunResource, nil
}

func (rn *ResourceNode) initThreeWayDiffData(operation *opsmodels.Operation) (*apiv1.Resource, *apiv1.Resource, *apiv1.Resource, v1.Status) {
	// 1. prepare planed resource that we want to execute
	planedResource := rn.resource
	// When a resource is deleted in Intent but exists in PriorState,
	// this node should be regarded as a deleted node, and rn.resource stores the PriorState
	if rn.Action == opsmodels.Delete {
		planedResource = nil
	}

	// 2. get prior resource which is stored in kusion_state.json
	key := rn.resource.ResourceKey()
	priorResource := operation.PriorStateResourceIndex[key]

	// 3. get the live resource from runtime
	readRequest := &runtime.ReadRequest{
		PlanResource:  planedResource,
		PriorResource: priorResource,
		Stack:         operation.Stack,
	}
	resourceType := rn.resource.Type
	response := operation.RuntimeMap[resourceType].Read(context.Background(), readRequest)
	liveResource := response.Resource
	s := response.Status
	if v1.IsErr(s) {
		return nil, nil, nil, s
	}
	return planedResource, priorResource, liveResource, nil
}

func removeNestedField(obj interface{}, fields ...string) {
	m := obj
	switch next := m.(type) {
	case map[string]interface{}:
		if len(fields) == 1 {
			delete(next, fields[0])
			return
		} else {
			removeNestedField(next[fields[0]], fields[1:]...)
		}
	case []interface{}:
		for _, n := range next {
			removeNestedField(n, fields...)
		}
	default:
		return
	}
}

func (rn *ResourceNode) applyResource(operation *opsmodels.Operation, prior, planed, live *apiv1.Resource) v1.Status {
	log.Infof("operation:%v, prior:%v, plan:%v, live:%v", rn.Action, jsonutil.Marshal2String(prior),
		jsonutil.Marshal2String(planed), jsonutil.Marshal2String(live))

	var res *apiv1.Resource
	var s v1.Status
	resourceType := rn.resource.Type

	rt := operation.RuntimeMap[resourceType]
	switch rn.Action {
	case opsmodels.Create, opsmodels.Update:
		response := rt.Apply(context.Background(), &runtime.ApplyRequest{PriorResource: prior, PlanResource: planed, Stack: operation.Stack})
		res = response.Resource
		s = response.Status
		log.Debugf("apply resource:%s, response: %v", planed.ID, jsonutil.Marshal2String(response))
	case opsmodels.Delete:
		response := rt.Delete(context.Background(), &runtime.DeleteRequest{Resource: prior, Stack: operation.Stack})
		s = response.Status
		if s != nil {
			log.Debugf("delete resource:%s, resource: %v", prior.ID, s.String())
		}
	case opsmodels.UnChanged:
		log.Infof("planed resource and live resource are equal")
		// auto import resources exist in intent and live cluster but no recorded in kusion_state.json
		if prior == nil {
			response := rt.Import(context.Background(), &runtime.ImportRequest{PlanResource: planed})
			s = response.Status
			log.Debugf("import resource:%s, resource:%v", planed.ID, jsonutil.Marshal2String(s))
			res = response.Resource
		} else {
			res = prior
		}
	}
	if v1.IsErr(s) {
		return s
	}

	key := rn.resource.ResourceKey()
	if e := operation.RefreshResourceIndex(key, res, rn.Action); e != nil {
		return v1.NewErrorStatus(e)
	}
	if e := operation.UpdateState(operation.StateResourceIndex); e != nil {
		return v1.NewErrorStatus(e)
	}

	// print apply resource success msg
	log.Infof("apply resource success: %s", rn.resource.ResourceKey())
	return nil
}

func (rn *ResourceNode) State() *apiv1.Resource {
	return rn.resource
}

func NewResourceNode(key string, state *apiv1.Resource, action opsmodels.ActionType) (*ResourceNode, v1.Status) {
	node, s := NewBaseNode(key)
	if v1.IsErr(s) {
		return nil, s
	}
	return &ResourceNode{baseNode: node, Action: action, resource: state}, nil
}

// save change steps in DAG walking order so that we can preview a full applying list
func updateChangeOrder(ops *opsmodels.Operation, rn *ResourceNode, plan, live interface{}) {
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

func ReplaceSecretRef(v reflect.Value) ([]string, reflect.Value, v1.Status) {
	return ReplaceRef(v, nil, nil)
}

var MustImplicitReplaceFun = func(resourceIndex map[string]*apiv1.Resource, refPath string) (reflect.Value, v1.Status) {
	return implicitReplaceFun(true, resourceIndex, refPath)
}

var OptionalImplicitReplaceFun = func(resourceIndex map[string]*apiv1.Resource, refPath string) (reflect.Value, v1.Status) {
	return implicitReplaceFun(false, resourceIndex, refPath)
}

// implicitReplaceFun will replace implicit dependency references. If force is true, this function will return an error when replace references failed
var implicitReplaceFun = func(
	force bool,
	resourceIndex map[string]*apiv1.Resource,
	refPath string,
) (reflect.Value, v1.Status) {
	const Sep = "."
	split := strings.Split(refPath, Sep)
	key := split[0]
	priorState := resourceIndex[key]
	if priorState == nil {
		msg := fmt.Sprintf("can't find resource by key:%s when replacing %s", key, refPath)
		return reflect.Value{}, v1.NewErrorStatusWithMsg(v1.IllegalManifest, msg)
	}
	attributes := priorState.Attributes
	if attributes == nil {
		msg := fmt.Sprintf("attributes is nil in resource:%s", key)
		return reflect.Value{}, v1.NewErrorStatusWithMsg(v1.IllegalManifest, msg)
	}
	var valueMap interface{}
	valueMap = attributes
	if len(split) > 1 {
		split := split[1:]
		for _, k := range split {
			if valueMap.(map[string]interface{})[k] == nil {
				if force {
					// only throw errors when force replacing operations like apply
					msg := fmt.Sprintf("can't find specified value in resource:%s by ref:%s", key, refPath)
					return reflect.Value{}, v1.NewErrorStatusWithMsg(v1.IllegalManifest, msg)
				} else {
					break
				}
			}
			valueMap = valueMap.(map[string]interface{})[k]
		}
	}
	return reflect.ValueOf(valueMap), nil
}

func ReplaceImplicitRef(
	v reflect.Value,
	resourceIndex map[string]*apiv1.Resource,
	replaceFun func(map[string]*apiv1.Resource, string) (reflect.Value, v1.Status),
) ([]string, reflect.Value, v1.Status) {
	return ReplaceRef(v, resourceIndex, replaceFun)
}

func ReplaceRef(
	v reflect.Value,
	resourceIndex map[string]*apiv1.Resource,
	repImplDepFunc func(map[string]*apiv1.Resource, string) (reflect.Value, v1.Status),
) ([]string, reflect.Value, v1.Status) {
	var result []string
	if !v.IsValid() {
		return nil, v, v1.NewErrorStatusWithMsg(v1.InvalidArgument, "invalid implicit reference")
	}

	switch v.Type().Kind() {
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return nil, v, nil
		}
		return ReplaceRef(v.Elem(), resourceIndex, repImplDepFunc)
	case reflect.String:
		vStr := v.String()
		if repImplDepFunc != nil {
			if strings.HasPrefix(vStr, ImplicitRefPrefix) {
				ref := strings.TrimPrefix(vStr, ImplicitRefPrefix)
				util.CheckArgument(len(ref) > 0,
					fmt.Sprintf("illegal implicit ref:%s. Implicit ref format: %sresourceKey.attribute", ref, ImplicitRefPrefix))
				split := strings.Split(ref, ".")
				result = append(result, split[0])
				log.Infof("add implicit ref:%s", split[0])
				// replace ref with actual value
				tv, s := repImplDepFunc(resourceIndex, ref)
				if v1.IsErr(s) {
					return nil, v, s
				}
				v = tv
			}
		}
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return nil, v, nil
		}

		vs := reflect.MakeSlice(v.Type(), 0, 0)

		for i := 0; i < v.Len(); i++ {
			ref, tv, s := ReplaceRef(v.Index(i), resourceIndex, repImplDepFunc)
			if v1.IsErr(s) {
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
			ref, tv, s := ReplaceRef(iter.Value(), resourceIndex, repImplDepFunc)
			if v1.IsErr(s) {
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

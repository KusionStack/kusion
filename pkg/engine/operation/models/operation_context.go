package models

import (
	"fmt"
	"sync"
	"time"

	"github.com/jinzhu/copier"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/state"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util"
	"kusionstack.io/kusion/pkg/util/json"
)

// Operation is the base model for all operations
type Operation struct {
	// OperationType represents the OperationType of this operation
	OperationType OperationType

	// StateStorage represents the storage where state will be saved during this operation
	StateStorage state.Storage

	// CtxResourceIndex represents resources updated by this operation
	CtxResourceIndex map[string]*v1.Resource

	// PriorStateResourceIndex represents resource state saved during the last operation
	PriorStateResourceIndex map[string]*v1.Resource

	// StateResourceIndex represents resources that will be saved in state.Storage
	StateResourceIndex map[string]*v1.Resource

	// IgnoreFields will be ignored in preview stage
	IgnoreFields []string

	// ChangeOrder is resources' change order during this operation
	ChangeOrder *ChangeOrder

	// RuntimeMap contains all infrastructure runtimes involved this operation. The key of this map is the Runtime type
	RuntimeMap map[v1.Type]runtime.Runtime

	// Stack contains info about where this command is invoked
	Stack *v1.Stack

	// MsgCh is used to send operation status like Success, Failed or Skip to Kusion CTl,
	// and this message will be displayed in the terminal
	MsgCh chan Message

	// Lock is the operation-wide mutex
	Lock *sync.Mutex

	// ResultState is the final DeprecatedState build by this operation, and this DeprecatedState will be saved in the StateStorage
	ResultState *v1.DeprecatedState
}

type Message struct {
	ResourceID string   // ResourceNode.ID()
	OpResult   OpResult // Success/Failed/Skip
	OpErr      error    // Operate error detail
}

type Request struct {
	Project  *v1.Project `json:"project"`
	Stack    *v1.Stack   `json:"stack"`
	Operator string      `json:"operator"`
	Intent   *v1.Spec    `json:"intent"`
}

type OpResult string

// OpResult values
const (
	Success OpResult = "Success"
	Failed  OpResult = "Failed"
	Skip    OpResult = "Skip"
)

// RefreshResourceIndex refresh resources in CtxResourceIndex & StateResourceIndex
func (o *Operation) RefreshResourceIndex(resourceKey string, resource *v1.Resource, actionType ActionType) error {
	o.Lock.Lock()
	defer o.Lock.Unlock()

	switch actionType {
	case Delete:
		o.CtxResourceIndex[resourceKey] = nil
		o.StateResourceIndex[resourceKey] = nil
	case Create, Update, UnChanged:
		o.CtxResourceIndex[resourceKey] = resource
		o.StateResourceIndex[resourceKey] = resource
	default:
		panic("unsupported actionType:" + actionType.Ing())
	}
	return nil
}

func (o *Operation) InitStates(request *Request) (*v1.DeprecatedState, *v1.DeprecatedState) {
	priorState, err := o.StateStorage.Get()
	util.CheckNotError(err, fmt.Sprintf("get state failed with request: %v", json.Marshal2PrettyString(request)))
	if priorState == nil {
		log.Infof("can't find state with request: %v", json.Marshal2PrettyString(request))
		priorState = v1.NewState()
	}
	resultState := v1.NewState()
	resultState.Serial = priorState.Serial
	err = copier.Copy(resultState, request)
	util.CheckNotError(err, fmt.Sprintf("copy request to result DeprecatedState failed, request:%v", json.Marshal2PrettyString(request)))
	resultState.Stack = request.Stack.Name
	resultState.Project = request.Project.Name

	resultState.Resources = nil

	return priorState, resultState
}

func (o *Operation) UpdateState(resourceIndex map[string]*v1.Resource) error {
	o.Lock.Lock()
	defer o.Lock.Unlock()

	resultState := o.ResultState
	resultState.Serial += 1
	resultState.Resources = nil

	res := make([]v1.Resource, 0, len(resourceIndex))
	for key := range resourceIndex {
		// {key -> nil} represents Deleted action
		if resourceIndex[key] == nil {
			continue
		}
		res = append(res, *resourceIndex[key])
	}

	resultState.Resources = res
	now := time.Now()
	if resultState.CreateTime.IsZero() {
		resultState.CreateTime = now
	}
	resultState.ModifiedTime = now
	err := o.StateStorage.Apply(resultState)
	if err != nil {
		return fmt.Errorf("apply DeprecatedState failed. %w", err)
	}
	log.Infof("update DeprecatedState:%v success", resultState.ID)
	return nil
}

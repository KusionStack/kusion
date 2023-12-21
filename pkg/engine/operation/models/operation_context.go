package models

import (
	"fmt"
	"sync"

	"github.com/jinzhu/copier"

	"kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util"
	jsonutil "kusionstack.io/kusion/pkg/util/json"
)

// Operation is the base model for all operations
type Operation struct {
	// OperationType represents the OperationType of this operation
	OperationType OperationType

	// StateStorage represents the storage where state will be saved during this operation
	StateStorage states.StateStorage

	// CtxResourceIndex represents resources updated by this operation
	CtxResourceIndex map[string]*intent.Resource

	// PriorStateResourceIndex represents resource state saved during the last operation
	PriorStateResourceIndex map[string]*intent.Resource

	// StateResourceIndex represents resources that will be saved in states.StateStorage
	StateResourceIndex map[string]*intent.Resource

	// IgnoreFields will be ignored in preview stage
	IgnoreFields []string

	// ChangeOrder is resources' change order during this operation
	ChangeOrder *ChangeOrder

	// RuntimeMap contains all infrastructure runtimes involved this operation. The key of this map is the Runtime type
	RuntimeMap map[intent.Type]runtime.Runtime

	// Stack contains info about where this command is invoked
	Stack *v1.Stack

	// MsgCh is used to send operation status like Success, Failed or Skip to Kusion CTl,
	// and this message will be displayed in the terminal
	MsgCh chan Message

	// Lock is the operation-wide mutex
	Lock *sync.Mutex

	// ResultState is the final State build by this operation, and this State will be saved in the StateStorage
	ResultState *states.State
}

type Message struct {
	ResourceID string   // ResourceNode.ID()
	OpResult   OpResult // Success/Failed/Skip
	OpErr      error    // Operate error detail
}

type Request struct {
	Tenant   string         `json:"tenant"`
	Project  *v1.Project    `json:"project"`
	Stack    *v1.Stack      `json:"stack"`
	Cluster  string         `json:"cluster"`
	Operator string         `json:"operator"`
	Intent   *intent.Intent `json:"intent"`
}

type OpResult string

// OpResult values
const (
	Success OpResult = "Success"
	Failed  OpResult = "Failed"
	Skip    OpResult = "Skip"
)

// RefreshResourceIndex refresh resources in CtxResourceIndex & StateResourceIndex
func (o *Operation) RefreshResourceIndex(resourceKey string, resource *intent.Resource, actionType ActionType) error {
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

func (o *Operation) InitStates(request *Request) (*states.State, *states.State) {
	query := &states.StateQuery{
		Tenant:  request.Tenant,
		Stack:   request.Stack.Name,
		Project: request.Project.Name,
		Cluster: request.Cluster,
	}
	latestState, err := o.StateStorage.GetLatestState(query)
	util.CheckNotError(err, fmt.Sprintf("get the latest State failed with query: %v", jsonutil.Marshal2PrettyString(query)))
	if latestState == nil {
		log.Infof("can't find states with request: %v", jsonutil.Marshal2PrettyString(request))
		latestState = states.NewState()
	}
	resultState := states.NewState()
	resultState.Serial = latestState.Serial
	err = copier.Copy(resultState, request)
	util.CheckNotError(err, fmt.Sprintf("copy request to result State failed, request:%v", jsonutil.Marshal2PrettyString(request)))
	resultState.Stack = request.Stack.Name
	resultState.Project = request.Project.Name

	resultState.Resources = nil

	return latestState, resultState
}

func (o *Operation) UpdateState(resourceIndex map[string]*intent.Resource) error {
	o.Lock.Lock()
	defer o.Lock.Unlock()

	state := o.ResultState
	state.Serial += 1
	state.Resources = nil

	res := make([]intent.Resource, 0, len(resourceIndex))
	for key := range resourceIndex {
		// {key -> nil} represents Deleted action
		if resourceIndex[key] == nil {
			continue
		}
		res = append(res, *resourceIndex[key])
	}

	state.Resources = res
	err := o.StateStorage.Apply(state)
	if err != nil {
		return fmt.Errorf("apply State failed. %w", err)
	}
	log.Infof("update State:%v success", state.ID)
	return nil
}

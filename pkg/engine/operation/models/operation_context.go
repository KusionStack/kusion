package models

import (
	"fmt"
	"sync"

	"github.com/jinzhu/copier"

	"kusionstack.io/kusion/pkg/engine/models"
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
	CtxResourceIndex map[string]*models.Resource

	// PriorStateResourceIndex represents resource state saved during the last operation
	PriorStateResourceIndex map[string]*models.Resource

	// StateResourceIndex represents resources that will be saved in states.StateStorage
	StateResourceIndex map[string]*models.Resource

	// IgnoreFields will be ignored in preview stage
	IgnoreFields []string

	// ChangeOrder is resources' change order during this operation
	ChangeOrder *ChangeOrder

	// Runtime is the resource infrastructure runtime of this operation
	Runtime runtime.Runtime

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
	Tenant   string       `json:"tenant"`
	Project  string       `json:"project"`
	Stack    string       `json:"stack"`
	Cluster  string       `json:"cluster"`
	Operator string       `json:"operator"`
	Spec     *models.Spec `json:"spec"`
}

type OpResult string

// OpResult values
const (
	Success OpResult = "Success"
	Failed  OpResult = "Failed"
	Skip    OpResult = "Skip"
)

// RefreshResourceIndex refresh resources in CtxResourceIndex & StateResourceIndex
func (o *Operation) RefreshResourceIndex(resourceKey string, resource *models.Resource, actionType ActionType) error {
	o.Lock.Lock()
	defer o.Lock.Unlock()

	switch actionType {
	case Delete:
		o.CtxResourceIndex[resourceKey] = nil
		o.StateResourceIndex[resourceKey] = nil
	case Create, Update, UnChange:
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
		Stack:   request.Stack,
		Project: request.Project,
		Cluster: request.Cluster,
	}
	latestState, err := o.StateStorage.GetLatestState(
		query,
	)
	util.CheckNotError(err, fmt.Sprintf("get the latest State failed with query: %v", jsonutil.Marshal2PrettyString(query)))
	if latestState == nil {
		log.Infof("can't find states with request: %v", jsonutil.Marshal2PrettyString(request))
		latestState = states.NewState()
	}
	resultState := states.NewState()
	resultState.Serial = latestState.Serial
	err = copier.Copy(resultState, request)
	util.CheckNotError(err, fmt.Sprintf("copy request to result State failed, request:%v", jsonutil.Marshal2PrettyString(request)))
	resultState.Resources = nil

	return latestState, resultState
}

func (o *Operation) UpdateState(resourceIndex map[string]*models.Resource) error {
	o.Lock.Lock()
	defer o.Lock.Unlock()

	state := o.ResultState
	state.Serial += 1
	state.Resources = nil

	res := make([]models.Resource, 0, len(resourceIndex))
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

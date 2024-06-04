package models

import (
	"fmt"
	"sync"
	"time"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/log"
)

// Operation is the base model for all operations
type Operation struct {
	// OperationType represents the OperationType of this operation
	OperationType OperationType

	// ReleaseStorage represents the storage where state will be saved during this operation
	ReleaseStorage release.Storage

	// CtxResourceIndex represents resources updated by this operation
	CtxResourceIndex map[string]*apiv1.Resource

	// PriorStateResourceIndex represents resource state saved during the last operation
	PriorStateResourceIndex map[string]*apiv1.Resource

	// StateResourceIndex represents resources that will be saved in state.Storage
	StateResourceIndex map[string]*apiv1.Resource

	// IgnoreFields will be ignored in preview stage
	IgnoreFields []string

	// ChangeOrder is resources' change order during this operation
	ChangeOrder *ChangeOrder

	// RuntimeMap contains all infrastructure runtimes involved this operation. The key of this map is the Runtime type
	RuntimeMap map[apiv1.Type]runtime.Runtime

	// Stack contains info about where this command is invoked
	Stack *apiv1.Stack

	// MsgCh is used to send operation status like Success, Failed or Skip to Kusion CTl,
	// and this message will be displayed in the terminal
	MsgCh chan Message

	// WatchCh is used to send the resource IDs that are ready to be watched after sending or executing
	// the apply request.
	// Fixme: try to merge the WatchCh with the MsgCh.
	WatchCh chan string

	// Lock is the operation-wide mutex
	Lock *sync.Mutex

	// Release is the release updated in this operation, and saved in the ReleaseStorage
	Release *apiv1.Release
}

type Message struct {
	ResourceID string   // ResourceNode.ID()
	OpResult   OpResult // Success/Failed/Skip
	OpErr      error    // Operate error detail
}

type Request struct {
	Project *apiv1.Project
	Stack   *apiv1.Stack
}

type OpResult string

// OpResult values
const (
	Success OpResult = "Success"
	Failed  OpResult = "Failed"
	Skip    OpResult = "Skip"
)

// RefreshResourceIndex refresh resources in CtxResourceIndex & StateResourceIndex
func (o *Operation) RefreshResourceIndex(resourceKey string, resource *apiv1.Resource, actionType ActionType) error {
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

func (o *Operation) UpdateReleaseState() error {
	o.Lock.Lock()
	defer o.Lock.Unlock()

	res := make([]apiv1.Resource, 0, len(o.StateResourceIndex))
	for key := range o.StateResourceIndex {
		// {key -> nil} represents Deleted action
		if o.StateResourceIndex[key] == nil {
			continue
		}
		res = append(res, *o.StateResourceIndex[key])
	}

	o.Release.State.Resources = res
	o.Release.ModifiedTime = time.Now()

	err := o.ReleaseStorage.Update(o.Release)
	if err != nil {
		return fmt.Errorf("udpate release failed, %w", err)
	}
	log.Infof("update release succeeded, project %s, workspace %s, revision %d", o.Release.Project, o.Release.Workspace, o.Release.Revision)
	return nil
}

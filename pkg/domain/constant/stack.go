package constant

import (
	"errors"
	"fmt"
)

// StackState represents the state of a stack.
type StackState string

// These constants represent the possible states of a stack.
const (
	// The stack has not been synced with the remote runtime.
	StackStateUnSynced StackState = "UnSynced"
	// The stack is synced with the remote runtime.
	StackStateSynced StackState = "Synced"
	// The stack has out of sync from the remote runtime.
	StackStateOutOfSync StackState = "OutOfSync"
)

var (
	ErrStackNil               = errors.New("stack is nil")
	ErrStackName              = errors.New("stack must have a name")
	ErrStackPath              = errors.New("stack must have a path")
	ErrStackFrameworkType     = errors.New("stack must have a framework type")
	ErrStackDesiredVersion    = errors.New("stack must have a desired version")
	ErrStackSource            = errors.New("stack must have a source")
	ErrStackSourceProvider    = errors.New("stack source must have a source provider")
	ErrStackRemote            = errors.New("stack source must have a remote")
	ErrStackSyncState         = errors.New("stack must have a sync state")
	ErrStackLastSyncTimestamp = errors.New("stack must have a last sync timestamp")
	ErrStackCreationTimestamp = errors.New("stack must have a creation timestamp")
	ErrStackUpdateTimestamp   = errors.New("stack must have a update timestamp")
	ErrStackHasNilProject     = errors.New("stack must have a project")
)

// ParseStackState parses a string into a StackState.
// If the string is not a valid StackState, it returns an error.
func ParseStackState(s string) (StackState, error) {
	switch s {
	case string(StackStateUnSynced):
		return StackStateUnSynced, nil
	case string(StackStateSynced):
		return StackStateSynced, nil
	case string(StackStateOutOfSync):
		return StackStateOutOfSync, nil
	default:
		return StackState(""), fmt.Errorf("invalid StackState: %q", s)
	}
}

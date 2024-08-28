package constant

import (
	"errors"
	"fmt"
)

// StackState represents the state of a stack.
type StackState string

// StackType represents the type of a stack.
type StackType string

// These constants represent the possible states of a stack.
const (
	// StackStateUnSynced represents state of stack has not been synced with the remote runtime.
	StackStateUnSynced StackState = "UnSynced"
	// StackStateSynced represents state of stack is synced with the remote runtime.
	StackStateSynced StackState = "Synced"
	// StackStateOutOfSync represents state of stack has out of sync from the remote runtime.
	StackStateOutOfSync StackState = "OutOfSync"

	StackStateCreating         StackState = "Creating"
	StackStateGenerating       StackState = "Generating"
	StackStateGenerateFailed   StackState = "GenerateFailed"
	StackStateGenerated        StackState = "Generated"
	StackStatePreviewing       StackState = "Previewing"
	StackStatePreviewFailed    StackState = "PreviewFailed"
	StackStatePreviewed        StackState = "Previewed"
	StackStateApplying         StackState = "Applying"
	StackStateApplyFailed      StackState = "ApplyFailed"
	StackStateApplySucceeded   StackState = "ApplySucceeded"
	StackStateDestroying       StackState = "Destroying"
	StackStateDestroyFailed    StackState = "DestroyFailed"
	StackStateDestroySucceeded StackState = "DestroySucceeded"

	StackTypeGlobal StackType = "global"
	StackTypeCloud  StackType = "cloud"
	StackTypeTenant StackType = "tenant"
	StackTypeCell   StackType = "cell"
	StackTypeBase   StackType = "base"
	StackTypeMain   StackType = "main"

	FirstSemanticVersion = "1.0.0"
	BaseStackName        = "base"
)

var (
	ErrStackNil                  = errors.New("stack is nil")
	ErrStackName                 = errors.New("stack must have a name")
	ErrStackPath                 = errors.New("stack must have a path")
	ErrStackNilOrPathEmpty       = errors.New("stack is nil or path is empty")
	ErrStackTypeInvalid          = errors.New("stack type is invalid")
	ErrStackFrameworkType        = errors.New("stack must have a framework type")
	ErrStackDesiredVersion       = errors.New("stack must have a desired version")
	ErrStackSource               = errors.New("stack must have a source")
	ErrStackSourceProvider       = errors.New("stack source must have a source provider")
	ErrStackRemote               = errors.New("stack source must have a remote")
	ErrStackSyncState            = errors.New("stack must have a sync state")
	ErrStackLastAppliedTimestamp = errors.New("stack must have a last sync timestamp")
	ErrStackCreationTimestamp    = errors.New("stack must have a creation timestamp")
	ErrStackUpdateTimestamp      = errors.New("stack must have a update timestamp")
	ErrStackHasNilProject        = errors.New("stack must have a project")
	ErrStackAlreadyExists        = errors.New("stack already exists")
	ErrProjectNameOrIDRequired   = errors.New("either project name or project ID is required")
	ErrGettingNonExistingProject = errors.New("project does not exist")
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

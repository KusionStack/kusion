package constant

import (
	"fmt"
)

type (
	RunType   string
	RunStatus string
)

const (
	RunTypeGenerate     RunType   = "Generate"
	RunTypePreview      RunType   = "Preview"
	RunTypeApply        RunType   = "Apply"
	RunTypeDestroy      RunType   = "Destroy"
	RunStatusScheduling RunStatus = "Scheduling"
	RunStatusInProgress RunStatus = "InProgress"
	RunStatusFailed     RunStatus = "Failed"
	RunStatusSucceeded  RunStatus = "Succeeded"
	RunStatusCancelled  RunStatus = "Cancelled"
	RunStatusQueued     RunStatus = "Queued"
)

// ParseRunType parses a string into a RunType.
// If the string is not a valid RunType, it returns an error.
func ParseRunType(s string) (RunType, error) {
	switch s {
	case string(RunTypeGenerate):
		return RunTypeGenerate, nil
	case string(RunTypePreview):
		return RunTypePreview, nil
	case string(RunTypeApply):
		return RunTypeApply, nil
	case string(RunTypeDestroy):
		return RunTypeDestroy, nil
	default:
		return RunType(""), fmt.Errorf("invalid RunType: %q", s)
	}
}

// ParseRunStatus parses a string into a RunStatus.
// If the string is not a valid RunStatus, it returns an error.
func ParseRunStatus(s string) (RunStatus, error) {
	switch s {
	case string(RunStatusScheduling):
		return RunStatusScheduling, nil
	case string(RunStatusInProgress):
		return RunStatusInProgress, nil
	case string(RunStatusFailed):
		return RunStatusFailed, nil
	case string(RunStatusSucceeded):
		return RunStatusSucceeded, nil
	case string(RunStatusCancelled):
		return RunStatusCancelled, nil
	case string(RunStatusQueued):
		return RunStatusQueued, nil
	default:
		return RunStatus(""), fmt.Errorf("invalid RunType: %q", s)
	}
}

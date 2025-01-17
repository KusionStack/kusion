package constant

import "strings"

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
	RunResultFailed     string    = "{\"result\":\"Operation Failed\"}"
	RunResultCancelled  string    = "{\"result\":\"Operation Cancelled\"}"
)

// ParseRunType parses a string into a RunType.
// If the string is not a valid RunType, it returns an error.
func ParseRunType(s string) (RunType, error) {
	switch strings.ToLower(s) {
	case strings.ToLower(string(RunTypeGenerate)):
		return RunTypeGenerate, nil
	case strings.ToLower(string(RunTypePreview)):
		return RunTypePreview, nil
	case strings.ToLower(string(RunTypeApply)):
		return RunTypeApply, nil
	case strings.ToLower(string(RunTypeDestroy)):
		return RunTypeDestroy, nil
	default:
		return RunType(""), nil
	}
}

// ParseRunStatus parses a string into a RunStatus.
// If the string is not a valid RunStatus, it returns an error.
func ParseRunStatus(s string) (RunStatus, error) {
	switch strings.ToLower(s) {
	case strings.ToLower(string(RunStatusScheduling)):
		return RunStatusScheduling, nil
	case strings.ToLower(string(RunStatusInProgress)):
		return RunStatusInProgress, nil
	case strings.ToLower(string(RunStatusFailed)):
		return RunStatusFailed, nil
	case strings.ToLower(string(RunStatusSucceeded)):
		return RunStatusSucceeded, nil
	case strings.ToLower(string(RunStatusCancelled)):
		return RunStatusCancelled, nil
	case strings.ToLower(string(RunStatusQueued)):
		return RunStatusQueued, nil
	default:
		return RunStatus(""), nil
	}
}

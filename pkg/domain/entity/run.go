package entity

import (
	"fmt"
	"time"

	"kusionstack.io/kusion/pkg/domain/constant"
)

// Run represents the specific run, including type
// which should be a specific instance of the run provider.
type Run struct {
	// ID is the id of the run.
	ID uint `yaml:"id" json:"id"`
	// RunType is the type of the run provider.
	Type constant.RunType `yaml:"type" json:"type"`
	// Stack is the stack of the run.
	Stack *Stack `yaml:"stack" json:"stack"`
	// Workspace is the target workspace of the run.
	Workspace string `yaml:"workspace" json:"workspace"`
	// Status is the status of the run.
	Status constant.RunStatus `yaml:"status" json:"status"`
	// Result is the result of the run.
	Result string `yaml:"result" json:"result"`
	// Trace is the trace of the run.
	Trace string `yaml:"trace" json:"trace"`
	// Logs is the logs of the run.
	Logs string `yaml:"logs" json:"logs"`
	// CreationTimestamp is the timestamp of the created for the run.
	CreationTimestamp time.Time `yaml:"creationTimestamp,omitempty" json:"creationTimestamp,omitempty"`
	// UpdateTimestamp is the timestamp of the updated for the run.
	UpdateTimestamp time.Time `yaml:"updateTimestamp,omitempty" json:"updateTimestamp,omitempty"`
}

// RunResult represents the result of the run.
type RunResult struct {
	// ExitCode is the exit code of the run.
	ExitCode int `yaml:"exitCode" json:"exitCode"`
	// Message is the message of the run.
	Message string `yaml:"message" json:"message"`
	// Old is the old state of the run.
	Old string `yaml:"old" json:"old"`
	// New is the new state of the run.
	New string `yaml:"new" json:"new"`
}

type RunFilter struct {
	ProjectID  uint
	StackID    uint
	Workspace  string
	Type       []string
	Status     []string
	StartTime  time.Time
	EndTime    time.Time
	Pagination *Pagination
}

type RunListResult struct {
	Runs  []*Run
	Total int
}

// Validate checks if the run is valid.
// It returns an error if the run is not valid.
func (r *Run) Validate() error {
	if r == nil {
		return fmt.Errorf("run is nil")
	}

	if r.Type == "" {
		return fmt.Errorf("run must have a run type")
	}

	if r.Workspace == "" {
		return fmt.Errorf("run must have a target workspace")
	}

	return nil
}

func (r *Run) Summary() string {
	return fmt.Sprintf("[%s][%s]", string(r.Type), string(r.Status))
}

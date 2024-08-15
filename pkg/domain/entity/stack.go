package entity

import (
	"time"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/domain/constant"
)

// Stack represents the specific configuration stack
type Stack struct {
	// ID is the id of the stack.
	ID uint `yaml:"id" json:"id"`
	// Name is the name of the stack.
	Name string `yaml:"name" json:"name"`
	// Type is the type of the stack.
	Type string `yaml:"type" json:"type"`
	// DisplayName is the human-readable display nams.
	DisplayName string `yaml:"displayName,omitempty" json:"displayName,omitempty"`
	// Project is the project associated with the stack.
	Project *Project `yaml:"project" json:"project"`
	// Desired is the desired version of stack.
	DesiredVersion string `yaml:"desiredVersion" json:"desiredVersion"`
	// Description is a human-readable description of the stack.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Path is the relative path of the stack within the sourcs.
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
	// Labels are custom labels associated with the stack.
	Labels []string `yaml:"labels,omitempty" json:"labels,omitempty"`
	// Owners is a list of owners for the stack.
	Owners []string `yaml:"owners,omitempty" json:"owners,omitempty"`
	// SyncState is the current state of the stack.
	SyncState constant.StackState `yaml:"syncState" json:"syncState"`
	// LastGeneratedRevision is the spec ID of the last generate operation for the stack.
	LastGeneratedRevision string `yaml:"lastGeneratedRevision" json:"lastGeneratedRevision"`
	// LastPreviewedRevision is the spec ID of the last preview operation for the stack.
	LastPreviewedRevision string `yaml:"lastPreviewedRevision" json:"lastPreviewedRevision"`
	// LastAppliedRevision is the spec ID of the last apply operation for the stack.
	LastAppliedRevision string `yaml:"lastAppliedRevision" json:"lastAppliedRevision"`
	// LastAppliedTimestamp is the timestamp of the last apply operation for the stack.
	LastAppliedTimestamp time.Time `yaml:"lastAppliedTimestamp,omitempty" json:"lastAppliedTimestamp,omitempty"`
	// CreationTimestamp is the timestamp of the created for the stack.
	CreationTimestamp time.Time `yaml:"creationTimestamp,omitempty" json:"creationTimestamp,omitempty"`
	// UpdateTimestamp is the timestamp of the updated for the stack.
	UpdateTimestamp time.Time `yaml:"updateTimestamp,omitempty" json:"updateTimestamp,omitempty"`
}

type StackFilter struct {
	OrgID     uint
	ProjectID uint
	Path      string
}

// Validate checks if the stack is valid.
// It returns an error if the stack is not valid.
func (s *Stack) Validate() error {
	if s == nil {
		return constant.ErrStackNil
	}

	if s.Name == "" {
		return constant.ErrStackName
	}

	if s.Project == nil {
		return constant.ErrStackHasNilProject
	}

	if err := s.Project.Validate(); err != nil {
		return err
	}

	if s.Path == "" {
		return constant.ErrStackPath
	}

	if s.SyncState == "" {
		return constant.ErrStackSyncState
	}

	return nil
}

// Convert stack to core stack
func (s *Stack) ConvertToCore() *v1.Stack {
	return &v1.Stack{
		Name:        s.Name,
		Description: &s.Description,
		Path:        s.Path,
		Labels:      map[string]string{},
	}
}

func (s *Stack) StackInOperation() bool {
	if s.SyncState == constant.StackStateGenerating ||
		s.SyncState == constant.StackStatePreviewing ||
		s.SyncState == constant.StackStateApplying ||
		s.SyncState == constant.StackStateDestroying {
		return true
	}
	return false
}

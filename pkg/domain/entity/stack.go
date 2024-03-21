package entity

import (
	"time"

	"kusionstack.io/kusion/pkg/domain/constant"
)

// Stack represents the specific configuration stack
type Stack struct {
	// ID is the id of the stack.
	ID uint `yaml:"id" json:"id"`
	// Name is the name of the stack.
	Name string `yaml:"name" json:"name"`
	// DisplayName is the readability display nams.
	DisplayName string `yaml:"displayName,omitempty" json:"displayName,omitempty"`
	// Source is the configuration source associated with the stack.
	// Source *Source `yaml:"source" json:"source"`
	// Project is the project associated with the stack.
	Project *Project `yaml:"project" json:"project"`
	// Org is the org associated with the stack.
	// Organization *Organization `yaml:"organization" json:"organization"`
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
	// LastSyncTimestamp is the timestamp of the last sync operation for the stack.
	LastSyncTimestamp time.Time `yaml:"lastSyncTimestamp,omitempty" json:"lastSyncTimestamp,omitempty"`
	// CreationTimestamp is the timestamp of the created for the stack.
	CreationTimestamp time.Time `yaml:"creationTimestamp,omitempty" json:"creationTimestamp,omitempty"`
	// UpdateTimestamp is the timestamp of the updated for the stack.
	UpdateTimestamp time.Time `yaml:"updateTimestamp,omitempty" json:"updateTimestamp,omitempty"`
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

	// if s.Source == nil {
	// 	return constant.ErrStackSource
	// }

	// if err := s.Source.Validate(); err != nil {
	// 	return err
	// }

	if len(s.DesiredVersion) == 0 {
		return constant.ErrStackDesiredVersion
	}

	return nil
}

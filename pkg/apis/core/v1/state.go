package v1

import (
	"time"

	"kusionstack.io/kusion/pkg/version"
)

// State is a record of an operation's result. It is a mapping between resources in KCL and the actual infra
// resource and often used as a datasource for 3-way merge/diff in operations like Apply or Preview.
type State struct {
	// State ID
	ID int64 `json:"id" yaml:"id"`

	// Project name
	Project string `json:"project" yaml:"project"`

	// Stack name
	Stack string `json:"stack" yaml:"stack"`

	// Workspace name
	Workspace string `json:"workspace" yaml:"workspace"`

	// State version
	Version int `json:"version" yaml:"version"`

	// KusionVersion represents the Kusion's version when this State is created
	KusionVersion string `json:"kusionVersion" yaml:"kusionVersion"`

	// Serial is an auto-increase number that represents how many times this State is modified
	Serial uint64 `json:"serial" yaml:"serial"`

	// Operator represents the person who triggered this operation
	Operator string `json:"operator,omitempty" yaml:"operator,omitempty"`

	// Resources records all resources in this operation
	Resources Resources `json:"resources" yaml:"resources"`

	// CreateTime is the time State is created
	CreateTime time.Time `json:"createTime" yaml:"createTime"`

	// ModifiedTime is the time State is modified each time
	ModifiedTime time.Time `json:"modifiedTime,omitempty" yaml:"modifiedTime,omitempty"`
}

func NewState() *State {
	s := &State{
		KusionVersion: version.ReleaseVersion(),
		Version:       1,
		Resources:     []Resource{},
	}
	return s
}

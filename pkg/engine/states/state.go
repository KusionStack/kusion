package states

import (
	"time"

	"kusionstack.io/kusion/pkg/engine/models"

	"github.com/zclconf/go-cty/cty"

	"kusionstack.io/kusion/pkg/version"
)

// StateStorage represents the set of methods required for a State backend
type StateStorage interface {
	// ConfigSchema returns a description of the expected configuration
	// structure for the receiving backend.
	ConfigSchema() cty.Type

	// Configure uses the provided configuration to set configuration fields
	// within the backend.
	Configure(obj cty.Value) error

	// GetLatestState return nil if state not exists
	GetLatestState(query *StateQuery) (*State, error)

	// Apply means update this state if it already exists or create a new one
	Apply(state *State) error

	// Delete State by id
	Delete(id string) error
}

type StateQuery struct {
	// Tenant name
	Tenant string `json:"tenant"`

	// Stack name
	Stack string `json:"stack"`

	// Project name
	Project string `json:"project"`

	// Cluster name
	Cluster string `json:"cluster,omitempty"`
}

// State is a record of an operation's result. It is a mapping between resources in KCL and the actual infra resource and often used as a
// datasource for 3-way merge/diff in operations like Apply or Preview.
type State struct {
	// State ID
	ID int64 `json:"id" yaml:"id"`

	// Tenant is designed for multi-tenant scenario
	Tenant string `json:"tenant,omitempty" yaml:"tenant,omitempty"`

	// Project name
	Project string `json:"project" yaml:"project"`

	// Stack name
	Stack string `json:"stack" yaml:"stack"`

	// Cluster is a logical concept to separate states in one stack.
	Cluster string `json:"cluster,omitempty" yaml:"cluster,omitempty"`

	// State version
	Version int `json:"version" yaml:"version"`

	// KusionVersion represents the Kusion's version when this State is created
	KusionVersion string `json:"kusionVersion" yaml:"kusionVersion"`

	// Serial is an auto-increase number that represents how many times this State is modified
	Serial uint64 `json:"serial" yaml:"serial"`

	// Operator represents the person who triggered this operation
	Operator string `json:"operator,omitempty" yaml:"operator,omitempty"`

	// Resources records all resources in this operation
	Resources models.Resources `json:"resources" yaml:"resources"`

	// CreateTime is the time State is created
	CreateTime time.Time `json:"createTime" yaml:"createTime"`

	// ModifiedTime is the time State is modified each time
	ModifiedTime time.Time `json:"modifiedTime,omitempty" yaml:"modifiedTime"`
}

func NewState() *State {
	s := &State{
		KusionVersion: version.ReleaseVersion(),
		Version:       1,
		Resources:     []models.Resource{},
	}
	return s
}

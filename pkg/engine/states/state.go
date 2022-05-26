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
	Tenant  string `json:"tenant"`
	Stack   string `json:"stack"`
	Project string `json:"project"`
}

// State represent all resources state in one apply operation.
type State struct {
	ID            int64            `json:"id"`
	Tenant        string           `json:"tenant"`
	Stack         string           `json:"stack"`
	Project       string           `json:"project"`
	Version       int              `json:"version"`
	KusionVersion string           `json:"kusionVersion"`
	Serial        uint64           `json:"serial"`
	Operator      string           `json:"operator"`
	Resources     models.Resources `json:"resources"`
	CreatTime     time.Time        `json:"creatTime"`
	ModifiedTime  time.Time        `json:"modifiedTime,omitempty"`
}

func NewState() *State {
	s := &State{
		KusionVersion: version.ReleaseVersion(),
		Version:       1,
		Resources:     []models.Resource{},
	}
	return s
}

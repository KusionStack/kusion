package states

import (
	"time"

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
	Delete(id string) error
}

type StateQuery struct {
	Tenant  string `json:"tenant"`
	Stack   string `json:"stack"`
	Project string `json:"project"`
}

// State represent all resources state in one apply operation.
type State struct {
	ID            int64     `json:"id"`
	Tenant        string    `json:"tenant"`
	Stack         string    `json:"stack"`
	Project       string    `json:"project"`
	Version       int       `json:"version"`
	KusionVersion string    `json:"kusion_version"`
	Serial        uint64    `json:"serial"`
	Operator      string    `json:"operator"`
	Resources     Resources `json:"resources"`
	GmtCreate     time.Time `json:"gmt_create"`
	GmtModified   time.Time `json:"gmt_modified,omitempty"`
}

func NewState() *State {
	s := &State{
		KusionVersion: version.ReleaseVersion(),
		Version:       1,
		Resources:     []ResourceState{},
	}
	return s
}

type Mode string

const (
	Managed Mode = "managed"
	Drifted      = "drifted"
)

type Resources []ResourceState

type ResourceState struct {
	// ID is the unique key of this resource in the whole State. ApiVersion/Kind/Namespace/Name is an idiomatic way of Kubernetes resources.
	ID         string                 `json:"id"`
	Mode       Mode                   `json:"mode"`
	Attributes map[string]interface{} `json:"attributes"`
	DependsOn  []string               `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
}

func (r *ResourceState) ResourceKey() string {
	return r.ID
}

func (rs Resources) Index() map[string]*ResourceState {
	m := make(map[string]*ResourceState)
	for i := range rs {
		m[rs[i].ResourceKey()] = &rs[i]
	}
	return m
}

func (rs Resources) Len() int      { return len(rs) }
func (rs Resources) Swap(i, j int) { rs[i], rs[j] = rs[j], rs[i] }
func (rs Resources) Less(i, j int) bool {
	switch {
	case rs[i].Mode != rs[j].Mode:
		return rs[i].Mode < rs[j].Mode
	case rs[i].ID != rs[j].ID:
		return rs[i].ID < rs[j].ID
	default:
		return false
	}
}

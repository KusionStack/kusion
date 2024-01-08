package workload

type Secret struct {
	Type      string            `yaml:"type" json:"type"`
	Params    map[string]string `yaml:"params,omitempty" json:"params,omitempty"`
	Data      map[string]string `yaml:"data,omitempty" json:"data,omitempty"`
	Immutable bool              `yaml:"immutable,omitempty" json:"immutable,omitempty"`
}

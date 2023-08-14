package workload

type WorkloadType string

const (
	WorkloadTypeJob     = "Job"
	WorkloadTypeService = "Service"
)

type Workload struct {
	Type     WorkloadType `yaml:"type" json:"type"`
	*Service `yaml:",inline" json:",inline"`
	*Job     `yaml:",inline" json:",inline"`
}

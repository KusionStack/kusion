package v1

import v1 "k8s.io/api/core/v1"

// Patcher contains fields should be patched into the workload corresponding fields
type Patcher struct {
	// Environments represent the environment variables patched to all containers in the workload.
	Environments []v1.EnvVar `json:"environments" yaml:"environments"`
	// Labels represent the labels patched to both the workload and pod.
	Labels map[string]string `json:"labels" yaml:"labels"`
	// Annotations represent the annotations patched to both the workload and pod.
	Annotations map[string]string `json:"annotations" yaml:"annotations"`
}

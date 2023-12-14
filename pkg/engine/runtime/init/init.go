package init

import (
	"fmt"
	"reflect"

	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/status"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes/kubeops"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform"
)

var SupportRuntimes = map[intent.Type]InitFn{
	runtime.Kubernetes: kubernetes.NewKubernetesRuntime,
	runtime.Terraform:  terraform.NewTerraformRuntime,
}

// InitFn runtime init func
type InitFn func(resource *intent.Resource) (runtime.Runtime, error)

func Runtimes(resources intent.Resources) (map[intent.Type]runtime.Runtime, status.Status) {
	runtimesMap := map[intent.Type]runtime.Runtime{}
	if resources == nil {
		return runtimesMap, nil
	}
	if errStatus := validResources(resources); errStatus != nil {
		return nil, errStatus
	}

	for _, resource := range resources {
		rt := resource.Type
		if runtimesMap[rt] == nil {
			r, err := SupportRuntimes[rt](&resource)
			if err != nil {
				return nil, status.NewErrorStatus(fmt.Errorf("init %s runtime failed. %w", rt, err))
			}
			runtimesMap[rt] = r
		}
	}
	return runtimesMap, nil
}

func validResources(resources intent.Resources) status.Status {
	var kubeConfig string
	for _, resource := range resources {
		rt := resource.Type
		if rt == "" {
			return status.NewErrorStatusWithCode(status.IllegalManifest, fmt.Errorf("no resource type in resource: %v", resource.ID))
		}
		if SupportRuntimes[rt] == nil {
			return status.NewErrorStatusWithCode(status.IllegalManifest, fmt.Errorf("unknown resource type: %s. Currently supported resource types are: %v",
				rt, reflect.ValueOf(SupportRuntimes).MapKeys()))
		}
		if rt == intent.Kubernetes {
			config := kubeops.GetKubeConfig(&resource)
			if kubeConfig != "" && kubeConfig != config {
				return status.NewErrorStatusWithCode(status.IllegalManifest, fmt.Errorf("different kubeConfig in different resources"))
			}
			if kubeConfig == "" {
				kubeConfig = config
			}
		}
	}
	return nil
}

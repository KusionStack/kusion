package init

import (
	"fmt"
	"reflect"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes/kubeops"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform"
)

var SupportRuntimes = map[apiv1.Type]InitFn{
	runtime.Kubernetes: kubernetes.NewKubernetesRuntime,
	runtime.Terraform:  terraform.NewTerraformRuntime,
}

// InitFn runtime init func
type InitFn func(resource *apiv1.Resource) (runtime.Runtime, error)

func Runtimes(resources apiv1.Resources) (map[apiv1.Type]runtime.Runtime, v1.Status) {
	runtimesMap := map[apiv1.Type]runtime.Runtime{}
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
				return nil, v1.NewErrorStatus(fmt.Errorf("init %s runtime failed. %w", rt, err))
			}
			runtimesMap[rt] = r
		}
	}
	return runtimesMap, nil
}

func validResources(resources apiv1.Resources) v1.Status {
	var kubeConfig string
	for _, resource := range resources {
		rt := resource.Type
		if rt == "" {
			return v1.NewErrorStatusWithCode(v1.IllegalManifest, fmt.Errorf("no resource type in resource: %v", resource.ID))
		}
		if SupportRuntimes[rt] == nil {
			return v1.NewErrorStatusWithCode(v1.IllegalManifest, fmt.Errorf("unknown resource type: %s. Currently supported resource types are: %v",
				rt, reflect.ValueOf(SupportRuntimes).MapKeys()))
		}
		if rt == apiv1.Kubernetes {
			config := kubeops.GetKubeConfig(&resource)
			if kubeConfig != "" && kubeConfig != config {
				return v1.NewErrorStatusWithCode(v1.IllegalManifest, fmt.Errorf("different kubeConfig in different resources"))
			}
			if kubeConfig == "" {
				kubeConfig = config
			}
		}
	}
	return nil
}

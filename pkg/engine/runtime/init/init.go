package init

import (
	"fmt"
	"reflect"

	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform"
	"kusionstack.io/kusion/pkg/status"
)

var SupportRuntimes = map[models.Type]InitFn{
	runtime.Kubernetes: kubernetes.NewKubernetesRuntime,
	runtime.Terraform:  terraform.NewTerraformRuntime,
}

// InitFn runtime init func
type InitFn func() (runtime.Runtime, error)

func Runtimes(resources models.Resources) (map[models.Type]runtime.Runtime, status.Status) {
	runtimesMap := map[models.Type]runtime.Runtime{}
	if resources == nil {
		return runtimesMap, nil
	}

	for _, resource := range resources {
		rt := resource.Type
		if rt == "" {
			return nil, status.NewErrorStatusWithCode(status.IllegalManifest, fmt.Errorf("no resource type in resource: %v", resource.ID))
		}

		if SupportRuntimes[rt] == nil {
			return nil, status.NewErrorStatusWithCode(status.IllegalManifest, fmt.Errorf("unknow resource type: %s. Currently supported resource types are: %v",
				rt, reflect.ValueOf(SupportRuntimes).MapKeys()))
		} else if runtimesMap[rt] == nil {
			r, err := SupportRuntimes[rt]()
			if err != nil {
				return nil, status.NewErrorStatus(fmt.Errorf("init %s runtime failed", rt))
			}
			runtimesMap[rt] = r
		}
	}

	return runtimesMap, nil
}

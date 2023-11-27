package init

import (
	"fmt"
	"reflect"

	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/stack"
	"kusionstack.io/kusion/pkg/apis/status"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform"
)

var SupportRuntimes = map[intent.Type]InitFn{
	runtime.Kubernetes: kubernetes.NewKubernetesRuntime,
	runtime.Terraform:  terraform.NewTerraformRuntime,
}

// InitFn runtime init func
type InitFn func(stack *stack.Stack) (runtime.Runtime, error)

func Runtimes(resources intent.Resources, stack *stack.Stack) (map[intent.Type]runtime.Runtime, status.Status) {
	runtimesMap := map[intent.Type]runtime.Runtime{}
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
			r, err := SupportRuntimes[rt](stack)
			if err != nil {
				return nil, status.NewErrorStatus(fmt.Errorf("init %s runtime failed. %w", rt, err))
			}
			runtimesMap[rt] = r
		}
	}

	return runtimesMap, nil
}

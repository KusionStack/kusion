package init

import (
	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform"
)

func InitRuntime() map[models.Type]InitFn {
	runtimes := map[models.Type]InitFn{
		runtime.Kubernetes: runtime.NewKubernetesRuntime,
		runtime.Terraform:  terraform.NewTerraformRuntime,
	}
	return runtimes
}

// InitFn init Runtime
type InitFn func() (runtime.Runtime, error)

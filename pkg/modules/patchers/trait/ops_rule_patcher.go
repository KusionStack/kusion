package trait

import (
	appsv1 "k8s.io/api/apps/v1"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	workspaceapi "kusionstack.io/kusion/pkg/apis/workspace"
	"kusionstack.io/kusion/pkg/modules"
	modelsapp "kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/trait"
)

type opsRulePatcher struct {
	app           *modelsapp.AppConfiguration
	modulesConfig map[string]workspaceapi.GenericConfig
}

// NewOpsRulePatcherFunc returns a NewPatcherFunc.
func NewOpsRulePatcherFunc(app *modelsapp.AppConfiguration, modulesConfig map[string]workspaceapi.GenericConfig) modules.NewPatcherFunc {
	return func() (modules.Patcher, error) {
		return NewOpsRulePatcher(app, modulesConfig)
	}
}

// NewOpsRulePatcher returns a Patcher.
func NewOpsRulePatcher(app *modelsapp.AppConfiguration, modulesConfig map[string]workspaceapi.GenericConfig) (modules.Patcher, error) {
	return &opsRulePatcher{
		app:           app,
		modulesConfig: modulesConfig,
	}, nil
}

// Patch implements Patcher interface.
func (p *opsRulePatcher) Patch(resources map[string][]*apiv1.Resource) error {
	if p.app.OpsRule == nil && p.modulesConfig["opsRule"] == nil {
		return nil
	}

	return modules.PatchResource[appsv1.Deployment](resources, modules.GVKDeployment, func(deploy *appsv1.Deployment) error {
		maxUnavailable, err := trait.GetMaxUnavailable(p.app.OpsRule, p.modulesConfig)
		if err != nil {
			return err
		}
		deploy.Spec.Strategy.Type = appsv1.RollingUpdateDeploymentStrategyType
		deploy.Spec.Strategy.RollingUpdate = &appsv1.RollingUpdateDeployment{
			MaxUnavailable: &maxUnavailable,
		}
		return nil
	})
}

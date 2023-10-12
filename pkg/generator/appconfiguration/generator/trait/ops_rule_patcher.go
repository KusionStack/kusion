package trait

import (
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	modelsapp "kusionstack.io/kusion/pkg/models/appconfiguration"
)

type opsRulePatcher struct {
	app *modelsapp.AppConfiguration
}

// NewOpsRulePatcherFunc returns a NewPatcherFunc.
func NewOpsRulePatcherFunc(app *modelsapp.AppConfiguration) appconfiguration.NewPatcherFunc {
	return func() (appconfiguration.Patcher, error) {
		return NewOpsRulePatcher(app)
	}
}

// NewOpsRulePatcher returns a Patcher.
func NewOpsRulePatcher(app *modelsapp.AppConfiguration) (appconfiguration.Patcher, error) {
	return &opsRulePatcher{
		app: app,
	}, nil
}

// Patch implements Patcher interface.
func (p *opsRulePatcher) Patch(resources map[string][]*models.Resource) error {
	if p.app.OpsRule == nil {
		return nil
	}

	return appconfiguration.PatchResource[appsv1.Deployment](resources, appconfiguration.GVKDeployment, func(deploy *appsv1.Deployment) error {
		maxUnavailable := intstr.Parse(p.app.OpsRule.MaxUnavailable)
		deploy.Spec.Strategy.Type = appsv1.RollingUpdateDeploymentStrategyType
		deploy.Spec.Strategy.RollingUpdate = &appsv1.RollingUpdateDeployment{
			MaxUnavailable: &maxUnavailable,
		}
		return nil
	})
}

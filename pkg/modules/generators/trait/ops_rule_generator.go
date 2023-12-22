package trait

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kusionstack.io/kube-api/apps/v1alpha1"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	appmodule "kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/trait"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

type opsRuleGenerator struct {
	project       *apiv1.Project
	stack         *apiv1.Stack
	appName       string
	app           *appmodule.AppConfiguration
	modulesConfig map[string]apiv1.GenericConfig
}

func NewOpsRuleGenerator(
	project *apiv1.Project,
	stack *apiv1.Stack,
	appName string,
	app *appmodule.AppConfiguration,
	modulesConfig map[string]apiv1.GenericConfig,
) (modules.Generator, error) {
	return &opsRuleGenerator{
		project:       project,
		stack:         stack,
		appName:       appName,
		app:           app,
		modulesConfig: modulesConfig,
	}, nil
}

func NewOpsRuleGeneratorFunc(
	project *apiv1.Project,
	stack *apiv1.Stack,
	appName string,
	app *appmodule.AppConfiguration,
	modulesConfig map[string]apiv1.GenericConfig,
) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewOpsRuleGenerator(project, stack, appName, app, modulesConfig)
	}
}

func (g *opsRuleGenerator) Generate(spec *apiv1.Intent) error {
	// opsRule does not exist in AppConfig and workspace config
	if g.app.OpsRule == nil && g.modulesConfig[trait.OpsRuleConst] == nil {
		return nil
	}

	// Job does not support maxUnavailable
	if g.app.Workload.Header.Type == workload.TypeJob {
		return nil
	}

	if g.app.Workload.Service.Type == workload.TypeCollaset {
		maxUnavailable, err := trait.GetMaxUnavailable(g.app.OpsRule, g.modulesConfig)
		if err != nil {
			return err
		}
		resource := &v1alpha1.PodTransitionRule{
			TypeMeta: metav1.TypeMeta{
				APIVersion: v1alpha1.GroupVersion.String(),
				Kind:       "PodTransitionRule",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      modules.UniqueAppName(g.project.Name, g.stack.Name, g.appName),
				Namespace: g.project.Name,
			},
			Spec: v1alpha1.PodTransitionRuleSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: modules.UniqueAppLabels(g.project.Name, g.appName),
				},
				Rules: []v1alpha1.TransitionRule{
					{
						Name: "maxUnavailable",
						TransitionRuleDefinition: v1alpha1.TransitionRuleDefinition{
							AvailablePolicy: &v1alpha1.AvailableRule{
								MaxUnavailableValue: &maxUnavailable,
							},
						},
					},
				},
			},
		}
		return modules.AppendToIntent(apiv1.Kubernetes, modules.KubernetesResourceID(resource.TypeMeta, resource.ObjectMeta), spec, resource)
	}
	return nil
}

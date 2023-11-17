package trait

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"kusionstack.io/kube-api/apps/v1alpha1"

	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	appmodule "kusionstack.io/kusion/pkg/models/appconfiguration"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload"
	"kusionstack.io/kusion/pkg/projectstack"
)

type opsRuleGenerator struct {
	project *projectstack.Project
	stack   *projectstack.Stack
	appName string
	app     *appmodule.AppConfiguration
}

func NewOpsRuleGenerator(
	project *projectstack.Project,
	stack *projectstack.Stack,
	appName string,
	app *appmodule.AppConfiguration,
) (appconfiguration.Generator, error) {
	return &opsRuleGenerator{
		project: project,
		stack:   stack,
		appName: appName,
		app:     app,
	}, nil
}

func NewOpsRuleGeneratorFunc(
	project *projectstack.Project,
	stack *projectstack.Stack,
	appName string,
	app *appmodule.AppConfiguration,
) appconfiguration.NewGeneratorFunc {
	return func() (appconfiguration.Generator, error) {
		return NewOpsRuleGenerator(project, stack, appName, app)
	}
}

func (g *opsRuleGenerator) Generate(spec *models.Intent) error {
	if g.app.OpsRule == nil {
		return nil
	}

	if g.app.Workload.Header.Type != workload.TypeService {
		return nil
	}

	if g.app.Workload.Service.Type != workload.TypeCollaset {
		return nil
	}

	maxUnavailable := intstr.Parse(g.app.OpsRule.MaxUnavailable)
	resource := &v1alpha1.RuleSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.GroupVersion.String(),
			Kind:       "RuleSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appconfiguration.UniqueAppName(g.project.Name, g.stack.Name, g.appName),
			Namespace: g.project.Name,
		},
		Spec: v1alpha1.RuleSetSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: appconfiguration.UniqueAppLabels(g.project.Name, g.appName),
			},
			Rules: []v1alpha1.RuleSetRule{
				{
					Name: "maxUnavailable",
					RuleSetRuleDefinition: v1alpha1.RuleSetRuleDefinition{
						AvailablePolicy: &v1alpha1.AvailableRule{
							MaxUnavailableValue: &maxUnavailable,
						},
					},
				},
			},
		},
	}
	return appconfiguration.AppendToSpec(models.Kubernetes, appconfiguration.KubernetesResourceID(resource.TypeMeta, resource.ObjectMeta), spec, resource)
}

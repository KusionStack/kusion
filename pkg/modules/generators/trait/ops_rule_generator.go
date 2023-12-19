package trait

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"kusionstack.io/kube-api/apps/v1alpha1"

	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/project"
	"kusionstack.io/kusion/pkg/apis/stack"
	"kusionstack.io/kusion/pkg/modules"
	appmodule "kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

type opsRuleGenerator struct {
	project *project.Project
	stack   *stack.Stack
	appName string
	app     *appmodule.AppConfiguration
}

func NewOpsRuleGenerator(
	project *project.Project,
	stack *stack.Stack,
	appName string,
	app *appmodule.AppConfiguration,
) (modules.Generator, error) {
	return &opsRuleGenerator{
		project: project,
		stack:   stack,
		appName: appName,
		app:     app,
	}, nil
}

func NewOpsRuleGeneratorFunc(
	project *project.Project,
	stack *stack.Stack,
	appName string,
	app *appmodule.AppConfiguration,
) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewOpsRuleGenerator(project, stack, appName, app)
	}
}

func (g *opsRuleGenerator) Generate(spec *intent.Intent) error {
	if g.app.OpsRule == nil {
		return nil
	}

	// Job does not support maxUnavailable
	if g.app.Workload.Header.Type == workload.TypeJob {
		return nil
	}

	if g.app.Workload.Service.Type == workload.TypeCollaset {
		maxUnavailable := intstr.Parse(g.app.OpsRule.MaxUnavailable)
		resource := &v1alpha1.RuleSet{
			TypeMeta: metav1.TypeMeta{
				APIVersion: v1alpha1.GroupVersion.String(),
				Kind:       "RuleSet",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      modules.UniqueAppName(g.project.Name, g.stack.Name, g.appName),
				Namespace: g.project.Name,
			},
			Spec: v1alpha1.RuleSetSpec{
				Selector: metav1.LabelSelector{
					MatchLabels: modules.UniqueAppLabels(g.project.Name, g.appName),
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
		return modules.AppendToIntent(intent.Kubernetes, modules.KubernetesResourceID(resource.TypeMeta, resource.ObjectMeta), spec, resource)
	}
	return nil
}

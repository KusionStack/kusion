package app_configuration

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration"
	"kusionstack.io/kusion/pkg/projectstack"
)

type AppConfigurationGenerator struct {
	*appconfiguration.AppConfiguration
}

func (acg *AppConfigurationGenerator) GenerateSpec(
	o *generator.Options,
	project *projectstack.Project,
	stack *projectstack.Stack,
) (*models.Spec, error) {
	resources := make(models.Resources, 0)

	if acg.AppConfiguration == nil {
		return nil, fmt.Errorf("can not find app configuration when generating the Spec")
	}

	if acg.AppConfiguration.Components != nil {

		// namespace
		ns := &v1.Namespace{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Namespace",
				APIVersion: v1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{Name: project.Name},
		}
		unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(ns)
		if err != nil {
			return nil, err
		}

		r := models.Resource{
			ID:         buildK8sResourceId(ns.TypeMeta, ns.ObjectMeta),
			Type:       generator.Kubernetes,
			Attributes: unstructured,
			DependsOn:  nil,
			Extensions: nil,
		}
		resources = append(resources, r)

		// Other resources
	}

	spec := &models.Spec{Resources: resources}
	return spec, nil
}

package app_configuration

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/projectstack"
)

// AppConfiguration is a developer-centric definition that describes how to run an Application.
// This application model builds upon a decade of experience at AntGroup running super large scale
// internal developer platform, combined with best-of-breed ideas and practices from the community.
//
// Example
//
//	components:
//	  proxy:
//	    containers:
//	      nginx:
//	        image: nginx:v1
//	        command:
//	        - /bin/sh
//	        - -c
//	        - echo hi
//	        args:
//	        - /bin/sh
//	        - -c
//	        - echo hi
//	        env:
//	          env1: VALUE
//	          env2: secret://sec-name/key
//	        workingDir: /tmp
//	    replicas: 2
type AppConfiguration struct {
	Components map[string]any
}

type AppConfigurationGenerator struct {
	AppConfiguration *AppConfiguration
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

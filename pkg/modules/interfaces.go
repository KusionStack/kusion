package modules

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules/inputs"
)

// GVKDeployment is the GroupVersionKind of Deployment
var (
	GVKDeployment = appsv1.SchemeGroupVersion.WithKind("Deployment").String()
	GVKService    = corev1.SchemeGroupVersion.WithKind("Service").String()
)

// Generator is an interface for things that can generate Intent from input
// configurations.
type Generator interface {
	// Generate performs the intent generate operation.
	Generate(intent *v1.Intent) error
}

// Patcher is the interface that wraps the Patch method.
type Patcher interface {
	Patch(resources map[string][]*v1.Resource) error
}

// NewGeneratorFunc is a function that returns a Generator.
type NewGeneratorFunc func() (Generator, error)

// NewPatcherFunc is a function that returns a Patcher.
type NewPatcherFunc func() (Patcher, error)

// GeneratorContext defines the context object used for generator.
type GeneratorContext struct {
	// Project provides basic project information for a given generator.
	Project *v1.Project

	// Stack provides basic stack information for a given generator.
	Stack *v1.Stack

	// Application provides basic application information for a given generator.
	Application *inputs.AppConfiguration

	// Namespace specifies the target Kubernetes namespace.
	Namespace string

	// ModuleInputs is the collection of module inputs for the target project.
	ModuleInputs map[string]v1.GenericConfig

	// TerraformConfig is the collection of provider configs for the terraform runtime.
	TerraformConfig v1.TerraformConfig

	// SecretStore is the external secret store spec
	SecretStore *v1.SecretStoreSpec
}

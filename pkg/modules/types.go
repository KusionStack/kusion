package modules

import (
	appsv1 "k8s.io/api/apps/v1"

	"kusionstack.io/kusion/pkg/apis/core/v1"
)

// GVKDeployment is the GroupVersionKind of Deployment
var GVKDeployment = appsv1.SchemeGroupVersion.WithKind("Deployment").String()

// Generator is the interface that wraps the Generate method.
type Generator interface {
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

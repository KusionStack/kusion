package orderedresources

import (
	"context"

	"github.com/jinzhu/copier"
	apiv1 "kusionstack.io/kusion-api-go/api.kusion.io/v1"
	modgen "kusionstack.io/kusion-module-framework/pkg/module/generator"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/generators"
)

// orderedResourcesGenerator is a generator that inject the dependsOn of resources in a specified order.
type orderedResourcesGenerator struct {
	orderedKinds []string
}

// NewOrderedResourcesGenerator returns a new instance of orderedResourcesGenerator.
func NewOrderedResourcesGenerator(multipleOrderedKinds ...[]string) (generators.SpecGenerator, error) {
	orderedKinds := modgen.DefaultOrderedKinds
	if len(multipleOrderedKinds) > 0 && len(multipleOrderedKinds[0]) > 0 {
		orderedKinds = multipleOrderedKinds[0]
	}
	return &orderedResourcesGenerator{
		orderedKinds: orderedKinds,
	}, nil
}

// NewOrderedResourcesGeneratorFunc returns a function that creates a new orderedResourcesGenerator.
func NewOrderedResourcesGeneratorFunc(multipleOrderedKinds ...[]string) generators.NewSpecGeneratorFunc {
	return func() (generators.SpecGenerator, error) {
		return NewOrderedResourcesGenerator(multipleOrderedKinds...)
	}
}

// Generate inject the dependsOn of resources in a specified order.
func (g *orderedResourcesGenerator) Generate(itt *v1.Spec) error {
	if itt.Resources == nil {
		itt.Resources = make(v1.Resources, 0)
	}

	// In Kusion, the type of `Resources` being passed around is internally defined,
	// so here we are converting the `Resources` type in `kusionstack.io/kusion-api-go` into
	// the internally defined type.
	// Fixme: the types passed around in Kusion may also be unified to the types in `kusion-api-go`.
	var resources apiv1.Resources
	copier.Copy(&resources, &itt.Resources)

	// Generate the ordered resources.
	orderedResources, err := modgen.OrderedResources(
		context.Background(),
		resources,
		g.orderedKinds,
	)
	if err != nil {
		return err
	}

	copier.Copy(&itt.Resources, &orderedResources)

	return nil
}

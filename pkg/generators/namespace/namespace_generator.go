package namespace

import (
	"context"

	"github.com/jinzhu/copier"
	modgen "kusionstack.io/kusion-module-framework/pkg/module/generator"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/generators"
)

type namespaceGenerator struct {
	namespace string
}

func NewNamespaceGenerator(namespace string) (generators.SpecGenerator, error) {
	return &namespaceGenerator{
		namespace: namespace,
	}, nil
}

func NewNamespaceGeneratorFunc(namespace string) generators.NewSpecGeneratorFunc {
	return func() (generators.SpecGenerator, error) {
		return NewNamespaceGenerator(namespace)
	}
}

func (g *namespaceGenerator) Generate(i *v1.Spec) error {
	if i.Resources == nil {
		i.Resources = make(v1.Resources, 0)
	}

	// Generate Kubernetes Namespace resource.
	ns, err := modgen.NamespaceResource(context.Background(), g.namespace)
	if err != nil {
		return err
	}

	// Avoid generating duplicate namespaces with the same ID.
	for _, res := range i.Resources {
		if res.ID == ns.ID {
			return nil
		}
	}

	// In Kusion, the type of `Resource` being passed around is internally defined,
	// so here we are converting the `Resource` type in `kusionstack.io/kusion-api-go` into
	// the internally defined type.
	// Fixme: the types passed around in Kusion may also be unified to the types in `kusion-api-go`.
	var res v1.Resource
	copier.Copy(&res, ns)

	i.Resources = append(i.Resources, res)

	return nil
}

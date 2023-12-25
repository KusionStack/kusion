package generators

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
)

type namespaceGenerator struct {
	context modules.GeneratorContext
}

func NewNamespaceGenerator(ctx modules.GeneratorContext) (modules.Generator, error) {
	return &namespaceGenerator{
		context: ctx,
	}, nil
}

func NewNamespaceGeneratorFunc(ctx modules.GeneratorContext) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewNamespaceGenerator(ctx)
	}
}

func (g *namespaceGenerator) Generate(i *apiv1.Intent) error {
	if i.Resources == nil {
		i.Resources = make(apiv1.Resources, 0)
	}

	ns := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{Name: g.context.Namespace},
	}

	// Avoid generating duplicate namespaces with the same ID.
	id := modules.KubernetesResourceID(ns.TypeMeta, ns.ObjectMeta)
	for _, res := range i.Resources {
		if res.ID == id {
			return nil
		}
	}

	return modules.AppendToIntent(apiv1.Kubernetes, id, i, ns)
}

package generators

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kusionstack.io/kusion/pkg/models"
)

func buildK8sResourceID(typeMeta metav1.TypeMeta, objectMeta metav1.ObjectMeta) string {
	// resource id example: apps/v1:Deployment:code-city:code-citydev
	id := typeMeta.APIVersion + ":" + typeMeta.Kind + ":"
	if objectMeta.Namespace != "" {
		id += objectMeta.Namespace + ":"
	}
	id += objectMeta.Name
	return id
}

func callGeneratorFuncs(newGenerators ...NewGeneratorFunc) ([]Generator, error) {
	gs := make([]Generator, 0, len(newGenerators))
	for _, newGenerator := range newGenerators {
		if g, err := newGenerator(); err != nil {
			return nil, err
		} else {
			gs = append(gs, g)
		}
	}

	return gs, nil
}

func callGenerators(spec *models.Spec, newGenerators ...NewGeneratorFunc) error {
	gs, err := callGeneratorFuncs(newGenerators...)
	if err != nil {
		return err
	}

	for _, g := range gs {
		if err := g.Generate(spec); err != nil {
			return err
		}
	}

	return nil
}

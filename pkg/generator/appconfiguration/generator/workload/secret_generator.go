package workload

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload"
	"kusionstack.io/kusion/pkg/projectstack"
)

type secretGenerator struct {
	project *projectstack.Project
	secrets map[string]workload.Secret
	appName string
}

func NewSecretGenerator(
	project *projectstack.Project,
	secrets map[string]workload.Secret,
	appName string,
) (appconfiguration.Generator, error) {
	if len(project.Name) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	if len(appName) == 0 {
		return nil, fmt.Errorf("app name must not be empty")
	}
	return &secretGenerator{
		project: project,
		secrets: secrets,
		appName: appName,
	}, nil
}

func NewSecretGeneratorFunc(
	project *projectstack.Project,
	secrets map[string]workload.Secret,
	appName string,
) appconfiguration.NewGeneratorFunc {
	return func() (appconfiguration.Generator, error) {
		return NewSecretGenerator(project, secrets, appName)
	}
}

func (g *secretGenerator) Generate(spec *models.Intent) error {
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	for sk, sv := range g.secrets {
		byteMap := make(map[string][]byte)
		for k, v := range sv.Data {
			byteMap[k] = []byte(v)
		}
		secret := &v1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: v1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{Name: sk, Namespace: g.project.Name},
			Data:       byteMap,
			Type:       sv.Type,
			Immutable:  &sv.Immutable,
		}

		err := appconfiguration.AppendToSpec(
			models.Kubernetes,
			appconfiguration.KubernetesResourceID(secret.TypeMeta, secret.ObjectMeta),
			spec,
			secret,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

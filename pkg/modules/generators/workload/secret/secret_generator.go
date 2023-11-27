package secret

import (
	"fmt"

	"golang.org/x/exp/maps"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/project"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

type secretGenerator struct {
	project *project.Project
	secrets map[string]workload.Secret
}

func NewSecretGenerator(
	project *project.Project,
	secrets map[string]workload.Secret,
) (modules.Generator, error) {
	if len(project.Name) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	return &secretGenerator{
		project: project,
		secrets: secrets,
	}, nil
}

func NewSecretGeneratorFunc(
	project *project.Project,
	secrets map[string]workload.Secret,
) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewSecretGenerator(project, secrets)
	}
}

func (g *secretGenerator) Generate(spec *intent.Intent) error {
	if spec.Resources == nil {
		spec.Resources = make(intent.Resources, 0)
	}

	for secretName, secretRef := range g.secrets {
		secret, err := generateSecret(g.project, secretName, secretRef)
		if err != nil {
			return err
		}

		resourceID := modules.KubernetesResourceID(secret.TypeMeta, secret.ObjectMeta)
		err = modules.AppendToSpec(
			intent.Kubernetes,
			resourceID,
			spec,
			secret,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// generateSecret generates target secret based on secret type. Most of these secret types are just semantic wrapper
// of native Kubernetes secret types:https://kubernetes.io/docs/concepts/configuration/secret/#secret-types, and more
// detailed usage info can be found in public documentation.
func generateSecret(project *project.Project, secretName string, secretRef workload.Secret) (*v1.Secret, error) {
	switch secretRef.Type {
	case "basic":
		return generateBasic(project, secretName, secretRef)
	case "token":
		return generateToken(project, secretName, secretRef)
	case "opaque":
		return generateOpaque(project, secretName, secretRef)
	case "certificate":
		return generateCertificate(project, secretName, secretRef)
	default:
		return nil, fmt.Errorf("unrecognized secret type %s", secretRef.Type)
	}
}

func generateBasic(project *project.Project, secretName string, secretRef workload.Secret) (*v1.Secret, error) {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: project.Name,
		},
		Data:      grabData(secretRef.Data, v1.BasicAuthUsernameKey, v1.BasicAuthPasswordKey),
		Immutable: &secretRef.Immutable,
		Type:      v1.SecretTypeBasicAuth,
	}

	for _, key := range []string{v1.BasicAuthUsernameKey, v1.BasicAuthPasswordKey} {
		if len(secret.Data[key]) == 0 {
			v := GenerateRandomString(54)
			secret.Data[key] = []byte(v)
		}
	}

	return secret, nil
}

func generateToken(project *project.Project, secretName string, secretRef workload.Secret) (*v1.Secret, error) {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: project.Name,
		},
		Data:      grabData(secretRef.Data, "token"),
		Immutable: &secretRef.Immutable,
		Type:      v1.SecretTypeOpaque,
	}

	if len(secret.Data["token"]) == 0 {
		v := GenerateRandomString(54)
		secret.Data["token"] = []byte(v)
	}

	return secret, nil
}

func generateOpaque(project *project.Project, secretName string, secretRef workload.Secret) (*v1.Secret, error) {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: project.Name,
		},
		Data:      grabData(secretRef.Data, maps.Keys(secretRef.Data)...),
		Immutable: &secretRef.Immutable,
		Type:      v1.SecretTypeOpaque,
	}

	return secret, nil
}

func generateCertificate(project *project.Project, secretName string, secretRef workload.Secret) (*v1.Secret, error) {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: project.Name,
		},
		Data:      grabData(secretRef.Data, v1.TLSCertKey, v1.TLSPrivateKeyKey),
		Immutable: &secretRef.Immutable,
		Type:      v1.SecretTypeTLS,
	}

	return secret, nil
}

// grabData extracts keys mapping data from original string map.
func grabData(from map[string]string, keys ...string) map[string][]byte {
	to := map[string][]byte{}
	for _, key := range keys {
		if v, ok := from[key]; ok {
			// don't override a non-zero length value with zero length
			if len(v) > 0 || len(to[key]) == 0 {
				to[key] = []byte(v)
			}
		}
	}
	return to
}

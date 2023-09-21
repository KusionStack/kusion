package workload

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"

	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload"
	"kusionstack.io/kusion/pkg/projectstack"
)

type Fields struct {
	project *projectstack.Project
	secrets map[string]workload.Secret
	appName string
}

type Args struct {
	spec *models.Spec
}

type TestCase struct {
	name    string
	fields  Fields
	args    Args
	want    *models.Spec
	wantErr bool
}

func BuildSecretTestCase(
	projectName, appName, secretName string,
	secretType v1.SecretType,
	secretData map[string]string,
	immutable bool,
) *TestCase {
	secretDataBase64 := map[string]interface{}{}
	for k, v := range secretData {
		secretDataBase64[k] = base64.StdEncoding.EncodeToString([]byte(v))
	}
	fmt.Println(secretDataBase64)
	expectedResources := []models.Resource{
		{
			ID:   fmt.Sprintf("v1:Secret:%s:%s", projectName, secretName),
			Type: "Kubernetes",
			Attributes: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"creationTimestamp": nil,
					"name":              secretName,
					"namespace":         projectName,
				},
				"data":      secretDataBase64,
				"immutable": immutable,
				"type":      string(secretType),
			},
			DependsOn:  nil,
			Extensions: nil,
		},
	}
	testCase := &TestCase{
		name: fmt.Sprintf("%s-%s-%s", projectName, appName, secretName),
		fields: Fields{
			project: &projectstack.Project{
				ProjectConfiguration: projectstack.ProjectConfiguration{
					Name: projectName,
				},
			},
			secrets: map[string]workload.Secret{
				secretName: {
					Type:      secretType,
					Data:      secretData,
					Immutable: immutable,
				},
			},
			appName: appName,
		},
		args: Args{
			spec: &models.Spec{},
		},
		want: &models.Spec{
			Resources: expectedResources,
		},
		wantErr: false,
	}
	return testCase
}

func TestSecretGenerator_Generate(t *testing.T) {
	tests := []TestCase{
		*BuildSecretTestCase("test-project", "test-app", "my-secret", v1.SecretTypeOpaque, map[string]string{"hello": "world", "mama": "miya", "foo": "bar"}, true),
		*BuildSecretTestCase("test-project", "test-app", "my-secret", v1.SecretTypeOpaque, map[string]string{"hello": "world", "mama": "miya", "foo": "bar"}, false),
		*BuildSecretTestCase("test-project", "test-app", "my-basic-secret", v1.SecretTypeBasicAuth, map[string]string{"username": "my-username", "password": "my-password"}, true),
		*BuildSecretTestCase("test-project", "test-app", "my-basic-secret", v1.SecretTypeBasicAuth, map[string]string{"username": "my-username", "password": "my-password"}, false),
		// test config with no secrets
		{
			fields: Fields{
				secrets: nil,
			},
			args: Args{
				spec: &models.Spec{},
			},
			want: &models.Spec{
				Resources: models.Resources{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &secretGenerator{
				project: tt.fields.project,
				secrets: tt.fields.secrets,
				appName: tt.fields.appName,
			}
			if err := g.Generate(tt.args.spec); (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.Equal(t, tt.want, tt.args.spec)
		})
	}
}

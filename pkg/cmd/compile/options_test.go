package compile

import (
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bytedance/mockey"
	"kusionstack.io/kusion/pkg/cmd/spec"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/projectstack"
)

var (
	apiVersion = "v1"
	kind       = "ServiceAccount"
	namespace  = "test-ns"

	project = &projectstack.Project{
		ProjectConfiguration: projectstack.ProjectConfiguration{
			Name:   "testdata",
			Tenant: "admin",
		},
	}
	stack = &projectstack.Stack{
		StackConfiguration: projectstack.StackConfiguration{
			Name: "dev",
		},
	}

	sa1 = newSA("sa1")
	sa2 = newSA("sa2")
	sa3 = newSA("sa3")
)

func TestCompileOptions_preSet(t *testing.T) {
	type fields struct {
		Settings []string
		Output   string
	}
	type want struct {
		Settings []string
		Output   string
	}

	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "preset-nothing",
			fields: fields{
				Settings: []string{"ci-test/settings.yaml", "kcl.yaml"},
				Output:   "ci-test/stdout.golden.yaml",
			},
			want: want{
				Settings: []string{"ci-test/settings.yaml", "kcl.yaml"},
				Output:   "ci-test/stdout.golden.yaml",
			},
		},
		{
			name: "preset-everything",
			fields: fields{
				Settings: []string{},
				Output:   "",
			},
			want: want{
				Settings: []string{"kcl.yaml"},
				Output:   "ci-test/stdout.golden.yaml",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewCompileOptions()

			o.Settings = tt.fields.Settings
			o.Output = tt.fields.Output

			o.PreSet(func(cur string) bool {
				return true
			})

			wantOpt := NewCompileOptions()
			wantOpt.Settings = tt.want.Settings
			wantOpt.Output = tt.want.Output

			assert.Equal(t, wantOpt, o)
		})
	}
}

func TestCompileOptions_Run(t *testing.T) {
	defer func() {
		os.Remove("kusion_state.json")
	}()

	t.Run("no style is true", func(t *testing.T) {
		m1 := mockDetectProjectAndStack()
		m2 := mockGenerateSpec()
		m3 := mockWriteFile()
		defer m1.UnPatch()
		defer m2.UnPatch()
		defer m3.UnPatch()

		o := NewCompileOptions()
		o.NoStyle = true
		err := o.Run()
		assert.Nil(t, err)
	})
}

func newSA(name string) models.Resource {
	return models.Resource{
		ID:   engine.BuildID(apiVersion, kind, namespace, name),
		Type: "Kubernetes",
		Attributes: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
		},
	}
}

func mockDetectProjectAndStack() *mockey.Mocker {
	return mockey.Mock(projectstack.DetectProjectAndStack).To(func(stackDir string) (*projectstack.Project, *projectstack.Stack, error) {
		project.Path = stackDir
		stack.Path = stackDir
		return project, stack, nil
	}).Build()
}

func mockGenerateSpec() *mockey.Mocker {
	return mockey.Mock(spec.GenerateSpecWithSpinner).To(func(
		o *generator.Options,
		project *projectstack.Project,
		stack *projectstack.Stack,
	) (*models.Spec, error) {
		return &models.Spec{Resources: []models.Resource{sa1, sa2, sa3}}, nil
	}).Build()
}

func mockWriteFile() *mockey.Mocker {
	return mockey.Mock(os.WriteFile).To(func(name string, data []byte, perm fs.FileMode) error {
		return nil
	}).Build()
}

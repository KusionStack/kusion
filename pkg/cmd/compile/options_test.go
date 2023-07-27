package compile

import (
	"io/fs"
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/cmd/spec"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/generator"
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

	want := NewCompileOptions()
	want.Settings = []string{"ci-test/settings.yaml", "kcl.yaml"}
	want.Output = "ci-test/stdout.golden.yaml"

	tests := []struct {
		name   string
		fields fields
		want   *CompileOptions
	}{
		{
			name: "preset-noting",
			fields: fields{
				Settings: []string{"ci-test/settings.yaml", "kcl.yaml"},
				Output:   "ci-test/stdout.golden.yaml",
			},
			want: want,
		},
		{
			name: "preset-everything",
			fields: fields{
				Settings: []string{},
				Output:   "",
			},
			want: want,
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
			assert.Equal(t, tt.want, o)
		})
	}
}

func TestCompileOptions_Run(t *testing.T) {
	defer func() {
		os.Remove("kusion_state.json")
	}()

	t.Run("no style is true", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockDetectProjectAndStack()
		mockGenerateSpec()
		mockWriteFile()

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

func mockDetectProjectAndStack() {
	monkey.Patch(projectstack.DetectProjectAndStack, func(stackDir string) (*projectstack.Project, *projectstack.Stack, error) {
		project.Path = stackDir
		stack.Path = stackDir
		return project, stack, nil
	})
}

func mockGenerateSpec() {
	monkey.Patch(spec.GenerateSpecWithSpinner, func(
		o *generator.Options,
		project *projectstack.Project,
		stack *projectstack.Stack,
	) (*models.Spec, error) {
		return &models.Spec{Resources: []models.Resource{sa1, sa2, sa3}}, nil
	})
}

func mockWriteFile() {
	monkey.Patch(os.WriteFile, func(name string, data []byte, perm fs.FileMode) error {
		return nil
	})
}

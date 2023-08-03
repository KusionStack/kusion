package compile

import (
	"errors"
	"github.com/bytedance/mockey"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"kusionstack.io/kusion/pkg/cmd/spec"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/models"
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

	testError = errors.New("test error")
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
		mockey.PatchConvey(tt.name, t, func() {
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

	mockey.PatchConvey("no style is true", t, func() {
		mockDetectProjectAndStack()
		mockGenerateSpec()
		mockWriteFile()

		o := NewCompileOptions()
		o.NoStyle = true
		err := o.Run()
		assert.Nil(t, err)
	})

	mockey.PatchConvey("detect project and spec failed", t, func() {
		mockDetectProjectAndStackFail()

		o := NewCompileOptions()
		o.NoStyle = true
		err := o.Run()
		assert.Equal(t, testError, err)
	})

	mockey.PatchConvey("generate spec failed", t, func() {
		mockDetectProjectAndStack()
		mockGenerateSpecFail()

		o := NewCompileOptions()
		o.NoStyle = true
		err := o.Run()
		assert.Equal(t, testError, err)
	})

	mockey.PatchConvey("write file failed", t, func() {
		mockDetectProjectAndStack()
		mockGenerateSpec()
		mockWriteFileFail()

		o := NewCompileOptions()
		o.NoStyle = true
		err := o.Run()
		assert.Equal(t, testError, err)
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
	mockey.Mock(projectstack.DetectProjectAndStack).To(func(stackDir string) (*projectstack.Project, *projectstack.Stack, error) {
		project.Path = stackDir
		stack.Path = stackDir
		return project, stack, nil
	}).Build()
}

func mockDetectProjectAndStackFail() {
	mockey.Mock(projectstack.DetectProjectAndStack).To(func(stackDir string) (*projectstack.Project, *projectstack.Stack, error) {
		project.Path = stackDir
		stack.Path = stackDir
		return project, stack, testError
	}).Build()
}

func mockGenerateSpec() {
	mockey.Mock(spec.GenerateSpecWithSpinner).To(func(o *generator.Options, project *projectstack.Project, stack *projectstack.Stack) (*models.Spec, error) {
		return &models.Spec{Resources: []models.Resource{sa1, sa2, sa3}}, nil
	}).Build()
}

func mockGenerateSpecFail() {
	mockey.Mock(spec.GenerateSpecWithSpinner).To(func(o *generator.Options, project *projectstack.Project, stack *projectstack.Stack) (*models.Spec, error) {
		return &models.Spec{Resources: []models.Resource{sa1, sa2, sa3}}, testError
	}).Build()
}

func mockWriteFile() {
	mockey.Mock(os.WriteFile).To(func(name string, data []byte, perm fs.FileMode) error {
		return nil
	}).Build()
}

func mockWriteFileFail() {
	mockey.Mock(os.WriteFile).To(func(name string, data []byte, perm fs.FileMode) error {
		return testError
	}).Build()
}

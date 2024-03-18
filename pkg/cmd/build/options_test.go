package build

import (
	"errors"
	"io/fs"
	"os"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

<<<<<<< HEAD
	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/cmd/build/builders"
=======
	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
>>>>>>> b551565 (feat: kusion server, engine api and refactor preview logic)
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/api/builders"
	"kusionstack.io/kusion/pkg/project"
	workspacestorages "kusionstack.io/kusion/pkg/workspace/storages"
)

var (
	apiVersion = "v1"
	kind       = "ServiceAccount"
	namespace  = "test-ns"

	proj = &v1.Project{
		Name: "testdata",
	}
	stack = &v1.Stack{
		Name: "dev",
	}

	sa1 = newSA("sa1")
	sa2 = newSA("sa2")
	sa3 = newSA("sa3")

	errTest = errors.New("test error")
)

func TestCompileOptions_preSet(t *testing.T) {
	type fields struct {
		Output string
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
			name: "preset-everything",
			fields: fields{
				Output: "",
			},
			want: want{
				Output:   "",
				Settings: []string{"kcl.yaml"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewBuildOptions()

			o.Output = tt.fields.Output
			_ = o.PreSet(func(cur string) bool {
				return true
			})

			wantOpt := NewBuildOptions()
			wantOpt.Output = tt.want.Output
			wantOpt.Settings = tt.want.Settings

			assert.Equal(t, wantOpt, o)
		})
	}
}

func TestBuildOptions_Run(t *testing.T) {
	t.Run("no style is true", func(t *testing.T) {
		mockey.PatchConvey("mock engine operation", t, func() {
			mockDetectProjectAndStack()
			mockWorkspaceStorage()
			mockGenerateIntent()
			mockWriteFile()

			o := NewBuildOptions()
			o.NoStyle = true
			err := o.Run()
			assert.Nil(t, err)
		})
	})

	t.Run("detect project and stack failed", func(t *testing.T) {
		mockey.PatchConvey("mock engine operation", t, func() {
			mockDetectProjectAndStackFail()

			o := NewBuildOptions()
			o.NoStyle = true
			err := o.Run()
			assert.Equal(t, errTest, err)
		})
	})

	t.Run("generate intent failed", func(t *testing.T) {
		mockey.PatchConvey("mock engine operation", t, func() {
			mockDetectProjectAndStack()
			mockWorkspaceStorage()
			mockGenerateIntentFail()

			o := NewBuildOptions()
			o.NoStyle = true
			err := o.Run()
			assert.Equal(t, errTest, err)
		})
	})
}

func newSA(name string) v1.Resource {
	return v1.Resource{
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
	mockey.Mock(project.DetectProjectAndStack).To(func(stackDir string) (*v1.Project, *v1.Stack, error) {
		proj.Path = stackDir
		stack.Path = stackDir
		return proj, stack, nil
	}).Build()
}

func mockDetectProjectAndStackFail() {
	mockey.Mock(project.DetectProjectAndStack).To(func(stackDir string) (*v1.Project, *v1.Stack, error) {
		proj.Path = stackDir
		stack.Path = stackDir
		return proj, stack, errTest
	}).Build()
}

func mockGenerateIntent() {
	mockey.Mock(IntentWithSpinner).To(func(
		o *builders.Options,
		proj *v1.Project,
		stack *v1.Stack,
		ws *v1.Workspace,
	) (*v1.Intent, error) {
		return &v1.Intent{Resources: []v1.Resource{sa1, sa2, sa3}}, nil
	}).Build()
}

func mockGenerateIntentFail() {
	mockey.Mock(IntentWithSpinner).To(func(
		o *builders.Options,
		proj *v1.Project,
		stack *v1.Stack,
		ws *v1.Workspace,
	) (*v1.Intent, error) {
		return &v1.Intent{Resources: []v1.Resource{sa1, sa2, sa3}}, errTest
	}).Build()
}

func mockWriteFile() {
	mockey.Mock(os.WriteFile).To(func(name string, data []byte, perm fs.FileMode) error {
		return nil
	}).Build()
}

func mockWorkspaceStorage() {
	mockey.Mock(backend.NewWorkspaceStorage).Return(&workspacestorages.LocalStorage{}, nil).Build()
	mockey.Mock((*workspacestorages.LocalStorage).Get).Return(&v1.Workspace{Name: "default"}, nil).Build()
}

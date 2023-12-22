package build

import (
	"errors"
	"io/fs"
	"os"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/cmd/build/builders"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/project"
)

var (
	apiVersion = "v1"
	kind       = "ServiceAccount"
	namespace  = "test-ns"

	p = &apiv1.Project{
		Name: "testdata",
	}
	s = &apiv1.Stack{
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
			o.PreSet(func(cur string) bool {
				return true
			})

			wantOpt := NewBuildOptions()
			wantOpt.Output = tt.want.Output
			wantOpt.Settings = tt.want.Settings

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
		m2 := mockGenerateIntent()
		m3 := mockWriteFile()
		defer m1.UnPatch()
		defer m2.UnPatch()
		defer m3.UnPatch()

		o := NewBuildOptions()
		o.NoStyle = true
		err := o.Run()
		assert.Nil(t, err)
	})

	mockey.PatchConvey("detect project and stack failed", t, func() {
		m1 := mockDetectProjectAndStackFail()
		defer m1.UnPatch()

		o := NewBuildOptions()
		o.NoStyle = true
		err := o.Run()
		assert.Equal(t, errTest, err)
	})

	mockey.PatchConvey("generate intent failed", t, func() {
		m1 := mockDetectProjectAndStack()
		m2 := mockGenerateIntentFail()
		defer m1.UnPatch()
		defer m2.UnPatch()
		o := NewBuildOptions()
		o.NoStyle = true
		err := o.Run()
		assert.Equal(t, errTest, err)
	})
}

func newSA(name string) apiv1.Resource {
	return apiv1.Resource{
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
	return mockey.Mock(project.DetectProjectAndStack).To(func(stackDir string) (*apiv1.Project, *apiv1.Stack, error) {
		p.Path = stackDir
		s.Path = stackDir
		return p, s, nil
	}).Build()
}

func mockDetectProjectAndStackFail() *mockey.Mocker {
	return mockey.Mock(project.DetectProjectAndStack).To(func(stackDir string) (*apiv1.Project, *apiv1.Stack, error) {
		p.Path = stackDir
		s.Path = stackDir
		return p, s, errTest
	}).Build()
}

func mockGenerateIntent() *mockey.Mocker {
	return mockey.Mock(IntentWithSpinner).To(func(
		o *builders.Options,
		project *apiv1.Project,
		stack *apiv1.Stack,
	) (*apiv1.Intent, error) {
		return &apiv1.Intent{Resources: []apiv1.Resource{sa1, sa2, sa3}}, nil
	}).Build()
}

func mockGenerateIntentFail() *mockey.Mocker {
	return mockey.Mock(IntentWithSpinner).To(func(
		o *builders.Options,
		project *apiv1.Project,
		stack *apiv1.Stack,
	) (*apiv1.Intent, error) {
		return &apiv1.Intent{Resources: []apiv1.Resource{sa1, sa2, sa3}}, errTest
	}).Build()
}

func mockWriteFile() *mockey.Mocker {
	return mockey.Mock(os.WriteFile).To(func(name string, data []byte, perm fs.FileMode) error {
		return nil
	}).Build()
}

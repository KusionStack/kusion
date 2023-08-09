package preview

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/require"

	compilecmd "kusionstack.io/kusion/pkg/cmd/compile"
	"kusionstack.io/kusion/pkg/cmd/spec"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/operation"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
	"kusionstack.io/kusion/pkg/engine/states/local"
	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/status"
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

func Test_preview(t *testing.T) {
	stateStorage := &local.FileSystemState{Path: filepath.Join("", local.KusionState)}
	t.Run("preview success", func(t *testing.T) {
		m := mockOperationPreview()
		defer m.UnPatch()

		o := NewPreviewOptions()
		_, err := Preview(o, stateStorage, &models.Spec{Resources: []models.Resource{sa1, sa2, sa3}}, project, stack)
		assert.Nil(t, err)
	})
}

func TestPreviewOptions_Run(t *testing.T) {
	defer func() {
		os.Remove("kusion_state.json")
	}()

	t.Run("no project or stack", func(t *testing.T) {
		o := NewPreviewOptions()
		o.Detail = true
		err := o.Run()
		assert.NotNil(t, err)
	})

	t.Run("compile failed", func(t *testing.T) {
		m := mockDetectProjectAndStack()
		defer m.UnPatch()

		o := NewPreviewOptions()
		o.Detail = true
		err := o.Run()
		assert.NotNil(t, err)
	})

	t.Run("no changes", func(t *testing.T) {
		m1 := mockDetectProjectAndStack()
		m2 := mockPatchGenerateSpecWithSpinner()
		m3 := mockNewKubernetesRuntime()
		defer m1.UnPatch()
		defer m2.UnPatch()
		defer m3.UnPatch()

		o := NewPreviewOptions()
		o.Detail = true
		err := o.Run()
		assert.Nil(t, err)
	})

	t.Run("detail is true", func(t *testing.T) {
		m1 := mockDetectProjectAndStack()
		m2 := mockPatchGenerateSpecWithSpinner()
		m3 := mockNewKubernetesRuntime()
		m4 := mockOperationPreview()
		m5 := mockPromptDetail("")
		defer m1.UnPatch()
		defer m2.UnPatch()
		defer m3.UnPatch()
		defer m4.UnPatch()
		defer m5.UnPatch()

		o := NewPreviewOptions()
		o.Detail = true
		err := o.Run()
		assert.Nil(t, err)
	})

	t.Run("json output is true", func(t *testing.T) {
		m1 := mockDetectProjectAndStack()
		m2 := mockGenerateSpec()
		m3 := mockNewKubernetesRuntime()
		m4 := mockOperationPreview()
		m5 := mockPromptDetail("")
		defer m1.UnPatch()
		defer m2.UnPatch()
		defer m3.UnPatch()
		defer m4.UnPatch()
		defer m5.UnPatch()

		o := NewPreviewOptions()
		o.Output = jsonOutput
		err := o.Run()
		assert.Nil(t, err)
	})

	t.Run("no style is true", func(t *testing.T) {
		m1 := mockDetectProjectAndStack()
		m2 := mockPatchGenerateSpecWithSpinner()
		m3 := mockNewKubernetesRuntime()
		m4 := mockOperationPreview()
		m5 := mockPromptDetail("")
		defer m1.UnPatch()
		defer m2.UnPatch()
		defer m3.UnPatch()
		defer m4.UnPatch()
		defer m5.UnPatch()

		o := NewPreviewOptions()
		o.NoStyle = true
		err := o.Run()
		assert.Nil(t, err)
	})
}

type fooRuntime struct{}

func (f *fooRuntime) Import(ctx context.Context, request *runtime.ImportRequest) *runtime.ImportResponse {
	return &runtime.ImportResponse{Resource: request.PlanResource}
}

func (f *fooRuntime) Apply(ctx context.Context, request *runtime.ApplyRequest) *runtime.ApplyResponse {
	return &runtime.ApplyResponse{
		Resource: request.PlanResource,
		Status:   nil,
	}
}

func (f *fooRuntime) Read(ctx context.Context, request *runtime.ReadRequest) *runtime.ReadResponse {
	if request.PlanResource.ResourceKey() == "fake-id" {
		return &runtime.ReadResponse{
			Resource: nil,
			Status:   nil,
		}
	}
	return &runtime.ReadResponse{
		Resource: request.PlanResource,
		Status:   nil,
	}
}

func (f *fooRuntime) Delete(ctx context.Context, request *runtime.DeleteRequest) *runtime.DeleteResponse {
	return nil
}

func (f *fooRuntime) Watch(ctx context.Context, request *runtime.WatchRequest) *runtime.WatchResponse {
	return nil
}

func mockOperationPreview() *mockey.Mocker {
	return mockey.Mock((*operation.PreviewOperation).Preview).To(func(
		*operation.PreviewOperation,
		*operation.PreviewRequest,
	) (rsp *operation.PreviewResponse, s status.Status) {
		return &operation.PreviewResponse{
			Order: &opsmodels.ChangeOrder{
				StepKeys: []string{sa1.ID, sa2.ID, sa3.ID},
				ChangeSteps: map[string]*opsmodels.ChangeStep{
					sa1.ID: {
						ID:     sa1.ID,
						Action: opsmodels.Create,
						From:   &sa1,
					},
					sa2.ID: {
						ID:     sa2.ID,
						Action: opsmodels.UnChanged,
						From:   &sa2,
					},
					sa3.ID: {
						ID:     sa3.ID,
						Action: opsmodels.Undefined,
						From:   &sa1,
					},
				},
			},
		}, nil
	}).Build()
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
	return mockey.Mock(spec.GenerateSpec).To(func(
		o *generator.Options,
		project *projectstack.Project,
		stack *projectstack.Stack,
	) (*models.Spec, error) {
		return &models.Spec{Resources: []models.Resource{sa1, sa2, sa3}}, nil
	}).Build()
}

func mockPatchGenerateSpecWithSpinner() *mockey.Mocker {
	return mockey.Mock(spec.GenerateSpecWithSpinner).To(func(
		o *generator.Options,
		project *projectstack.Project,
		stack *projectstack.Stack,
	) (*models.Spec, error) {
		return &models.Spec{Resources: []models.Resource{sa1, sa2, sa3}}, nil
	}).Build()
}

func mockNewKubernetesRuntime() *mockey.Mocker {
	return mockey.Mock(kubernetes.NewKubernetesRuntime).To(func() (runtime.Runtime, error) {
		return &fooRuntime{}, nil
	}).Build()
}

func mockPromptDetail(input string) *mockey.Mocker {
	return mockey.Mock((*opsmodels.ChangeOrder).PromptDetails).To(func(co *opsmodels.ChangeOrder) (string, error) {
		return input, nil
	}).Build()
}

func TestPreviewOptions_ValidateSpecFile(t *testing.T) {
	currDir, _ := os.Getwd()
	tests := []struct {
		name           string
		specFile       string
		workDir        string
		createSpecFile bool
		wantErr        bool
	}{
		{
			name:           "test1",
			specFile:       "kusion_spec.yaml",
			workDir:        "",
			createSpecFile: true,
		},
		{
			name:           "test2",
			specFile:       filepath.Join(currDir, "kusion_spec.yaml"),
			workDir:        "",
			createSpecFile: true,
		},
		{
			name:           "test3",
			specFile:       "kusion_spec.yaml",
			workDir:        "",
			createSpecFile: false,
			wantErr:        true,
		},
		{
			name:           "test4",
			specFile:       "ci-test/stdout.golden.yaml",
			workDir:        "",
			createSpecFile: true,
		},
		{
			name:           "test5",
			specFile:       "../kusion_spec.yaml",
			workDir:        "",
			createSpecFile: true,
			wantErr:        true,
		},
		{
			name:           "test6",
			specFile:       filepath.Join(currDir, "../kusion_spec.yaml"),
			workDir:        "",
			createSpecFile: true,
			wantErr:        true,
		},
		{
			name:     "test7",
			specFile: "",
			workDir:  "",
			wantErr:  false,
		},
		{
			name:     "test8",
			specFile: currDir,
			workDir:  "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := PreviewOptions{}
			o.SpecFile = tt.specFile
			o.WorkDir = tt.workDir
			if tt.createSpecFile {
				dir := filepath.Dir(tt.specFile)
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					os.MkdirAll(dir, 0o755)
					defer os.RemoveAll(dir)
				}
				os.Create(tt.specFile)
				defer os.Remove(tt.specFile)
			}
			err := o.ValidateSpecFile()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestPreviewOptions_Validate(t *testing.T) {
	m := mockey.Mock((*compilecmd.CompileOptions).Validate).Return(nil).Build()
	defer m.UnPatch()
	tests := []struct {
		name    string
		output  string
		wantErr bool
	}{
		{
			name:    "test1",
			output:  "json",
			wantErr: false,
		},
		{
			name:    "test2",
			output:  "yaml",
			wantErr: true,
		},
		{
			name:    "test3",
			output:  "",
			wantErr: false,
		},
		{
			name:    "test4",
			output:  "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &PreviewOptions{}
			o.Output = tt.output
			err := o.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

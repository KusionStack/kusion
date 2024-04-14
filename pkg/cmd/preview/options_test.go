package preview

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/backend/storages"
	"kusionstack.io/kusion/pkg/cmd/build"
	"kusionstack.io/kusion/pkg/cmd/generate"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
	statestorages "kusionstack.io/kusion/pkg/engine/state/storages"
	"kusionstack.io/kusion/pkg/project"
	workspacestorages "kusionstack.io/kusion/pkg/workspace/storages"
)

var (
	apiVersion = "v1"
	kind       = "ServiceAccount"
	namespace  = "test-ns"

	proj = &apiv1.Project{
		Name: "testdata",
	}
	stack = &apiv1.Stack{
		Name: "dev",
	}

	sa1 = newSA("sa1")
	sa2 = newSA("sa2")
	sa3 = newSA("sa3")
)

func Test_preview(t *testing.T) {
	stateStorage := statestorages.NewLocalStorage(filepath.Join("", "state.yaml"))
	t.Run("preview success", func(t *testing.T) {
		m := mockOperationPreview()
		defer m.UnPatch()

		o := NewPreviewOptions()
		_, err := Preview(o, stateStorage, &apiv1.Spec{Resources: []apiv1.Resource{sa1, sa2, sa3}}, proj, stack)
		assert.Nil(t, err)
	})
}

func TestPreviewOptions_Run(t *testing.T) {
	t.Run("no project or stack", func(t *testing.T) {
		o := NewPreviewOptions()
		o.Detail = true
		err := o.Run()
		assert.NotNil(t, err)
	})

	t.Run("compile failed", func(t *testing.T) {
		mockey.PatchConvey("mock engine operation", t, func() {
			mockDetectProjectAndStack()

			o := NewPreviewOptions()
			o.Detail = true
			err := o.Run()
			assert.NotNil(t, err)
		})
	})

	t.Run("no changes", func(t *testing.T) {
		mockey.PatchConvey("mock engine operation", t, func() {
			mockDetectProjectAndStack()
			mockGenerateIntentWithSpinner()
			mockNewKubernetesRuntime()
			mockNewBackend()
			mockWorkspaceStorage()
			o := NewPreviewOptions()
			o.Detail = true
			err := o.Run()
			assert.Nil(t, err)
		})
	})

	t.Run("detail is true", func(t *testing.T) {
		mockey.PatchConvey("mock engine operation", t, func() {
			mockDetectProjectAndStack()
			mockGenerateIntentWithSpinner()
			mockNewKubernetesRuntime()
			mockOperationPreview()
			mockPromptDetail("")
			mockNewBackend()
			mockWorkspaceStorage()

			o := NewPreviewOptions()
			o.Detail = true
			err := o.Run()
			assert.Nil(t, err)
		})
	})

	t.Run("json output is true", func(t *testing.T) {
		mockey.PatchConvey("mock engine operation", t, func() {
			mockDetectProjectAndStack()
			mockGenerateIntentWithSpinner()
			mockNewKubernetesRuntime()
			mockOperationPreview()
			mockPromptDetail("")
			mockNewBackend()
			mockWorkspaceStorage()

			o := NewPreviewOptions()
			o.Output = jsonOutput
			err := o.Run()
			assert.Nil(t, err)
		})
	})

	t.Run("no style is true", func(t *testing.T) {
		mockey.PatchConvey("mock engine operation", t, func() {
			mockDetectProjectAndStack()
			mockGenerateIntentWithSpinner()
			mockNewKubernetesRuntime()
			mockOperationPreview()
			mockPromptDetail("")
			mockNewBackend()
			mockWorkspaceStorage()

			o := NewPreviewOptions()
			o.NoStyle = true
			err := o.Run()
			assert.Nil(t, err)
		})
	})
}

type fooRuntime struct{}

func (f *fooRuntime) Import(_ context.Context, request *runtime.ImportRequest) *runtime.ImportResponse {
	return &runtime.ImportResponse{Resource: request.PlanResource}
}

func (f *fooRuntime) Apply(_ context.Context, request *runtime.ApplyRequest) *runtime.ApplyResponse {
	return &runtime.ApplyResponse{
		Resource: request.PlanResource,
		Status:   nil,
	}
}

func (f *fooRuntime) Read(_ context.Context, request *runtime.ReadRequest) *runtime.ReadResponse {
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

func (f *fooRuntime) Delete(_ context.Context, _ *runtime.DeleteRequest) *runtime.DeleteResponse {
	return nil
}

func (f *fooRuntime) Watch(_ context.Context, _ *runtime.WatchRequest) *runtime.WatchResponse {
	return nil
}

func mockOperationPreview() *mockey.Mocker {
	return mockey.Mock((*operation.PreviewOperation).Preview).To(func(
		*operation.PreviewOperation,
		*operation.PreviewRequest,
	) (rsp *operation.PreviewResponse, s v1.Status) {
		return &operation.PreviewResponse{
			Order: &models.ChangeOrder{
				StepKeys: []string{sa1.ID, sa2.ID, sa3.ID},
				ChangeSteps: map[string]*models.ChangeStep{
					sa1.ID: {
						ID:     sa1.ID,
						Action: models.Create,
						From:   &sa1,
					},
					sa2.ID: {
						ID:     sa2.ID,
						Action: models.UnChanged,
						From:   &sa2,
					},
					sa3.ID: {
						ID:     sa3.ID,
						Action: models.Undefined,
						From:   &sa1,
					},
				},
			},
		}, nil
	}).Build()
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

func mockDetectProjectAndStack() {
	mockey.Mock(project.DetectProjectAndStackFrom).To(func(stackDir string) (*apiv1.Project, *apiv1.Stack, error) {
		proj.Path = stackDir
		stack.Path = stackDir
		return proj, stack, nil
	}).Build()
}

func mockGenerateIntentWithSpinner() {
	mockey.Mock(generate.GenerateSpecWithSpinner).To(func(
		project *apiv1.Project,
		stack *apiv1.Stack,
		workspace *apiv1.Workspace,
		parameters map[string]string,
		noStyle bool,
	) (*apiv1.Spec, error) {
		return &apiv1.Spec{Resources: []apiv1.Resource{sa1, sa2, sa3}}, nil
	}).Build()
}

func mockNewKubernetesRuntime() {
	mockey.Mock(kubernetes.NewKubernetesRuntime).To(func() (runtime.Runtime, error) {
		return &fooRuntime{}, nil
	}).Build()
}

func mockPromptDetail(input string) {
	mockey.Mock((*models.ChangeOrder).PromptDetails).To(func(co *models.ChangeOrder) (string, error) {
		return input, nil
	}).Build()
}

func mockNewBackend() {
	mockey.Mock(backend.NewBackend).Return(&storages.LocalStorage{}, nil).Build()
}

func mockWorkspaceStorage() {
	mockey.Mock((*storages.LocalStorage).WorkspaceStorage).Return(&workspacestorages.LocalStorage{}, nil).Build()
	mockey.Mock((*workspacestorages.LocalStorage).Get).Return(&apiv1.Workspace{}, nil).Build()
}

func TestPreviewOptions_ValidateIntentFile(t *testing.T) {
	currDir, _ := os.Getwd()
	tests := []struct {
		name             string
		intentFile       string
		workDir          string
		createIntentFile bool
		wantErr          bool
	}{
		{
			name:             "test1",
			intentFile:       "kusion_intent.yaml",
			workDir:          "",
			createIntentFile: true,
		},
		{
			name:             "test2",
			intentFile:       filepath.Join(currDir, "kusion_intent.yaml"),
			workDir:          "",
			createIntentFile: true,
		},
		{
			name:             "test3",
			intentFile:       "kusion_intent.yaml",
			workDir:          "",
			createIntentFile: false,
			wantErr:          true,
		},
		{
			name:             "test4",
			intentFile:       "ci-test/stdout.golden.yaml",
			workDir:          "",
			createIntentFile: true,
		},
		{
			name:             "test5",
			intentFile:       "../kusion_intent.yaml",
			workDir:          "",
			createIntentFile: true,
			wantErr:          true,
		},
		{
			name:             "test6",
			intentFile:       filepath.Join(currDir, "../kusion_intent.yaml"),
			workDir:          "",
			createIntentFile: true,
			wantErr:          true,
		},
		{
			name:       "test7",
			intentFile: "",
			workDir:    "",
			wantErr:    false,
		},
		{
			name:       "test8",
			intentFile: currDir,
			workDir:    "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := Options{}
			o.IntentFile = tt.intentFile
			o.WorkDir = tt.workDir
			if tt.createIntentFile {
				dir := filepath.Dir(tt.intentFile)
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					_ = os.MkdirAll(dir, 0o755)
					defer func() {
						_ = os.RemoveAll(dir)
					}()
				}
				_, _ = os.Create(tt.intentFile)
				defer func() {
					_ = os.Remove(tt.intentFile)
				}()
			}
			err := o.ValidateIntentFile()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestPreviewOptions_Validate(t *testing.T) {
	m := mockey.Mock((*build.Options).Validate).Return(nil).Build()
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
			o := &Options{}
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

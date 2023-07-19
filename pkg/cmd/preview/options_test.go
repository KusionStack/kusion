package preview

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/cmd/spec"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/engine/operation"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
	"kusionstack.io/kusion/pkg/engine/states/local"
	"kusionstack.io/kusion/pkg/generator"
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
		defer monkey.UnpatchAll()
		mockOperationPreview()

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
		defer monkey.UnpatchAll()
		mockDetectProjectAndStack()

		o := NewPreviewOptions()
		o.Detail = true
		err := o.Run()
		assert.NotNil(t, err)
	})

	t.Run("no changes", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockDetectProjectAndStack()
		mockGenerateSpecWithSpinner()
		mockNewKubernetesRuntime()

		o := NewPreviewOptions()
		o.Detail = true
		err := o.Run()
		assert.Nil(t, err)
	})

	t.Run("detail is true", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockDetectProjectAndStack()
		mockGenerateSpecWithSpinner()
		mockNewKubernetesRuntime()
		mockOperationPreview()
		mockPromptDetail("")

		o := NewPreviewOptions()
		o.Detail = true
		err := o.Run()
		assert.Nil(t, err)
	})

	t.Run("json output is true", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockDetectProjectAndStack()
		mockGenerateSpec()
		mockNewKubernetesRuntime()
		mockOperationPreview()
		mockPromptDetail("")

		o := NewPreviewOptions()
		o.Output = jsonOutput
		err := o.Run()
		assert.Nil(t, err)
	})

	t.Run("no style is true", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockDetectProjectAndStack()
		mockGenerateSpecWithSpinner()
		mockNewKubernetesRuntime()
		mockOperationPreview()
		mockPromptDetail("")

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

func mockOperationPreview() {
	monkey.Patch((*operation.PreviewOperation).Preview,
		func(*operation.PreviewOperation, *operation.PreviewRequest) (rsp *operation.PreviewResponse, s status.Status) {
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
		},
	)
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

func mockGenerateSpecWithSpinner() {
	monkey.Patch(spec.GenerateSpecWithSpinner, func(
		o *generator.Options,
		project *projectstack.Project,
		stack *projectstack.Stack,
	) (*models.Spec, error) {
		return &models.Spec{Resources: []models.Resource{sa1, sa2, sa3}}, nil
	})
}

func mockGenerateSpec() {
	monkey.Patch(spec.GenerateSpec, func(
		o *generator.Options,
		project *projectstack.Project,
		stack *projectstack.Stack,
	) (*models.Spec, error) {
		return &models.Spec{Resources: []models.Resource{sa1, sa2, sa3}}, nil
	})
}

func mockNewKubernetesRuntime() {
	monkey.Patch(kubernetes.NewKubernetesRuntime, func() (runtime.Runtime, error) {
		return &fooRuntime{}, nil
	})
}

func mockPromptDetail(input string) {
	monkey.Patch((*opsmodels.ChangeOrder).PromptDetails, func(co *opsmodels.ChangeOrder) (string, error) {
		return input, nil
	})
}

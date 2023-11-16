package destroy

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/AlecAivazis/survey/v2"
	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/operation"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/engine/states/local"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/status"
)

func TestDestroyOptions_Run(t *testing.T) {
	mockey.PatchConvey("Detail is true", t, func() {
		mockDetectProjectAndStack()
		mockGetLatestState()
		mockNewKubernetesRuntime()
		mockOperationPreview()

		o := NewDestroyOptions()
		o.Detail = true
		err := o.Run()
		assert.Nil(t, err)
	})

	mockey.PatchConvey("prompt no", t, func() {
		mockDetectProjectAndStack()
		mockGetLatestState()
		mockNewKubernetesRuntime()
		mockOperationPreview()

		o := NewDestroyOptions()
		mockPromptOutput("no")
		err := o.Run()
		assert.Nil(t, err)
	})

	mockey.PatchConvey("prompt yes", t, func() {
		mockDetectProjectAndStack()
		mockGetLatestState()
		mockNewKubernetesRuntime()
		mockOperationPreview()
		mockOperationDestroy(opsmodels.Success)

		o := NewDestroyOptions()
		mockPromptOutput("yes")
		err := o.Run()
		assert.Nil(t, err)
	})
}

var (
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
)

func mockDetectProjectAndStack() {
	mockey.Mock(projectstack.DetectProjectAndStack).To(func(stackDir string) (*projectstack.Project, *projectstack.Stack, error) {
		project.Path = stackDir
		stack.Path = stackDir
		return project, stack, nil
	}).Build()
}

func mockGetLatestState() {
	mockey.Mock((*local.FileSystemState).GetLatestState).To(func(
		f *local.FileSystemState,
		query *states.StateQuery,
	) (*states.State, error) {
		return &states.State{Resources: []models.Resource{sa1}}, nil
	}).Build()
}

func Test_preview(t *testing.T) {
	mockey.PatchConvey("preview success", t, func() {
		mockNewKubernetesRuntime()
		mockOperationPreview()

		o := NewDestroyOptions()
		stateStorage := &local.FileSystemState{Path: filepath.Join(o.WorkDir, local.KusionState)}
		_, err := o.preview(&models.Intent{Resources: []models.Resource{sa1}}, project, stack, stateStorage)
		assert.Nil(t, err)
	})
}

func mockNewKubernetesRuntime() {
	mockey.Mock(kubernetes.NewKubernetesRuntime).To(func() (runtime.Runtime, error) {
		return &fakerRuntime{}, nil
	}).Build()
}

var _ runtime.Runtime = (*fakerRuntime)(nil)

type fakerRuntime struct{}

func (f *fakerRuntime) Import(ctx context.Context, request *runtime.ImportRequest) *runtime.ImportResponse {
	return &runtime.ImportResponse{Resource: request.PlanResource}
}

func (f *fakerRuntime) Apply(ctx context.Context, request *runtime.ApplyRequest) *runtime.ApplyResponse {
	return &runtime.ApplyResponse{
		Resource: request.PlanResource,
		Status:   nil,
	}
}

func (f *fakerRuntime) Read(ctx context.Context, request *runtime.ReadRequest) *runtime.ReadResponse {
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

func (f *fakerRuntime) Delete(ctx context.Context, request *runtime.DeleteRequest) *runtime.DeleteResponse {
	return nil
}

func (f *fakerRuntime) Watch(ctx context.Context, request *runtime.WatchRequest) *runtime.WatchResponse {
	return nil
}

func mockOperationPreview() {
	mockey.Mock((*operation.PreviewOperation).Preview).To(
		func(*operation.PreviewOperation, *operation.PreviewRequest) (rsp *operation.PreviewResponse, s status.Status) {
			return &operation.PreviewResponse{
				Order: &opsmodels.ChangeOrder{
					StepKeys: []string{sa1.ID},
					ChangeSteps: map[string]*opsmodels.ChangeStep{
						sa1.ID: {
							ID:     sa1.ID,
							Action: opsmodels.Delete,
							From:   nil,
						},
					},
				},
			}, nil
		},
	).Build()
}

const (
	apiVersion = "v1"
	kind       = "ServiceAccount"
	namespace  = "test-ns"
)

var (
	sa1 = newSA("sa1")
	sa2 = newSA("sa2")
)

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

func Test_destroy(t *testing.T) {
	mockey.PatchConvey("destroy success", t, func() {
		mockNewKubernetesRuntime()
		mockOperationDestroy(opsmodels.Success)

		o := NewDestroyOptions()
		planResources := &models.Intent{Resources: []models.Resource{sa2}}
		order := &opsmodels.ChangeOrder{
			StepKeys: []string{sa1.ID, sa2.ID},
			ChangeSteps: map[string]*opsmodels.ChangeStep{
				sa1.ID: {
					ID:     sa1.ID,
					Action: opsmodels.Delete,
					From:   nil,
				},
				sa2.ID: {
					ID:     sa2.ID,
					Action: opsmodels.UnChanged,
					From:   &sa2,
				},
			},
		}
		changes := opsmodels.NewChanges(project, stack, order)

		stateStorage := &local.FileSystemState{Path: filepath.Join(o.WorkDir, local.KusionState)}

		err := o.destroy(planResources, changes, stateStorage)
		assert.Nil(t, err)
	})
	mockey.PatchConvey("destroy failed", t, func() {
		mockNewKubernetesRuntime()
		mockOperationDestroy(opsmodels.Failed)

		o := NewDestroyOptions()
		planResources := &models.Intent{Resources: []models.Resource{sa1}}
		order := &opsmodels.ChangeOrder{
			StepKeys: []string{sa1.ID},
			ChangeSteps: map[string]*opsmodels.ChangeStep{
				sa1.ID: {
					ID:     sa1.ID,
					Action: opsmodels.Delete,
					From:   nil,
				},
			},
		}
		changes := opsmodels.NewChanges(project, stack, order)
		stateStorage := &local.FileSystemState{Path: filepath.Join(o.WorkDir, local.KusionState)}

		err := o.destroy(planResources, changes, stateStorage)
		assert.NotNil(t, err)
	})
}

func mockOperationDestroy(res opsmodels.OpResult) {
	mockey.Mock((*operation.DestroyOperation).Destroy).To(
		func(o *operation.DestroyOperation, request *operation.DestroyRequest) status.Status {
			var err error
			if res == opsmodels.Failed {
				err = errors.New("mock error")
			}
			for _, r := range request.Spec.Resources {
				// ing -> $res
				o.MsgCh <- opsmodels.Message{
					ResourceID: r.ResourceKey(),
					OpResult:   "",
					OpErr:      nil,
				}
				o.MsgCh <- opsmodels.Message{
					ResourceID: r.ResourceKey(),
					OpResult:   res,
					OpErr:      err,
				}
			}
			close(o.MsgCh)
			if res == opsmodels.Failed {
				return status.NewErrorStatus(err)
			}
			return nil
		}).Build()
}

func Test_prompt(t *testing.T) {
	mockey.PatchConvey("prompt error", t, func() {
		mockey.Mock(
			survey.AskOne).To(
			func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
				return errors.New("mock error")
			},
		).Build()
		_, err := prompt()
		assert.NotNil(t, err)
	})

	mockey.PatchConvey("prompt yes", t, func() {
		mockPromptOutput("yes")
		_, err := prompt()
		assert.Nil(t, err)
	})
}

func mockPromptOutput(res string) {
	mockey.Mock(
		survey.AskOne).To(
		func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			reflect.ValueOf(response).Elem().Set(reflect.ValueOf(res))
			return nil
		},
	).Build()
}

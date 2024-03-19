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

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/backend/storages"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
	statestorages "kusionstack.io/kusion/pkg/engine/state/storages"
	"kusionstack.io/kusion/pkg/project"
	workspacestorages "kusionstack.io/kusion/pkg/workspace/storages"
)

func TestDestroyOptions_Run(t *testing.T) {
	mockey.PatchConvey("Detail is true", t, func() {
		mockDetectProjectAndStack()
		mockGetState()
		mockBackend()
		mockWorkspaceStorage()
		mockNewKubernetesRuntime()
		mockOperationPreview()

		o := NewDestroyOptions()
		o.Detail = true
		err := o.Run()
		assert.Nil(t, err)
	})

	mockey.PatchConvey("prompt no", t, func() {
		mockDetectProjectAndStack()
		mockGetState()
		mockBackend()
		mockWorkspaceStorage()
		mockNewKubernetesRuntime()
		mockOperationPreview()

		o := NewDestroyOptions()
		mockPromptOutput("no")
		err := o.Run()
		assert.Nil(t, err)
	})

	mockey.PatchConvey("prompt yes", t, func() {
		mockDetectProjectAndStack()
		mockGetState()
		mockBackend()
		mockWorkspaceStorage()
		mockNewKubernetesRuntime()
		mockOperationPreview()
		mockOperationDestroy(models.Success)

		o := NewDestroyOptions()
		mockPromptOutput("yes")
		err := o.Run()
		assert.Nil(t, err)
	})
}

var (
	proj = &apiv1.Project{
		Name: "testdata",
	}
	stack = &apiv1.Stack{
		Name: "dev",
	}
)

func mockDetectProjectAndStack() {
	mockey.Mock(project.DetectProjectAndStack).To(func(stackDir string) (*apiv1.Project, *apiv1.Stack, error) {
		proj.Path = stackDir
		stack.Path = stackDir
		return proj, stack, nil
	}).Build()
}

func mockGetState() {
	mockey.Mock((*statestorages.LocalStorage).Get).Return(&apiv1.State{Resources: []apiv1.Resource{sa1}}, nil).Build()
}

func Test_preview(t *testing.T) {
	mockey.PatchConvey("preview success", t, func() {
		mockNewKubernetesRuntime()
		mockOperationPreview()

		o := NewDestroyOptions()
		stateStorage := statestorages.NewLocalStorage(filepath.Join(o.WorkDir, "state.yaml"))
		_, err := o.preview(&apiv1.Intent{Resources: []apiv1.Resource{sa1}}, proj, stack, stateStorage)
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

func (f *fakerRuntime) Import(_ context.Context, request *runtime.ImportRequest) *runtime.ImportResponse {
	return &runtime.ImportResponse{Resource: request.PlanResource}
}

func (f *fakerRuntime) Apply(_ context.Context, request *runtime.ApplyRequest) *runtime.ApplyResponse {
	return &runtime.ApplyResponse{
		Resource: request.PlanResource,
		Status:   nil,
	}
}

func (f *fakerRuntime) Read(_ context.Context, request *runtime.ReadRequest) *runtime.ReadResponse {
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

func (f *fakerRuntime) Delete(_ context.Context, _ *runtime.DeleteRequest) *runtime.DeleteResponse {
	return nil
}

func (f *fakerRuntime) Watch(_ context.Context, _ *runtime.WatchRequest) *runtime.WatchResponse {
	return nil
}

func mockOperationPreview() {
	mockey.Mock((*operation.PreviewOperation).Preview).To(
		func(*operation.PreviewOperation, *operation.PreviewRequest) (rsp *operation.PreviewResponse, s v1.Status) {
			return &operation.PreviewResponse{
				Order: &models.ChangeOrder{
					StepKeys: []string{sa1.ID},
					ChangeSteps: map[string]*models.ChangeStep{
						sa1.ID: {
							ID:     sa1.ID,
							Action: models.Delete,
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

func Test_destroy(t *testing.T) {
	mockey.PatchConvey("destroy success", t, func() {
		mockNewKubernetesRuntime()
		mockOperationDestroy(models.Success)

		o := NewDestroyOptions()
		planResources := &apiv1.Intent{Resources: []apiv1.Resource{sa2}}
		order := &models.ChangeOrder{
			StepKeys: []string{sa1.ID, sa2.ID},
			ChangeSteps: map[string]*models.ChangeStep{
				sa1.ID: {
					ID:     sa1.ID,
					Action: models.Delete,
					From:   nil,
				},
				sa2.ID: {
					ID:     sa2.ID,
					Action: models.UnChanged,
					From:   &sa2,
				},
			},
		}
		changes := models.NewChanges(proj, stack, order)

		stateStorage := statestorages.NewLocalStorage(filepath.Join(o.WorkDir, "state.yaml"))

		err := o.destroy(planResources, changes, stateStorage)
		assert.Nil(t, err)
	})
	mockey.PatchConvey("destroy failed", t, func() {
		mockNewKubernetesRuntime()
		mockOperationDestroy(models.Failed)

		o := NewDestroyOptions()
		planResources := &apiv1.Intent{Resources: []apiv1.Resource{sa1}}
		order := &models.ChangeOrder{
			StepKeys: []string{sa1.ID},
			ChangeSteps: map[string]*models.ChangeStep{
				sa1.ID: {
					ID:     sa1.ID,
					Action: models.Delete,
					From:   nil,
				},
			},
		}
		changes := models.NewChanges(proj, stack, order)
		stateStorage := statestorages.NewLocalStorage(filepath.Join(o.WorkDir, "state.yaml"))

		err := o.destroy(planResources, changes, stateStorage)
		assert.NotNil(t, err)
	})
}

func mockOperationDestroy(res models.OpResult) {
	mockey.Mock((*operation.DestroyOperation).Destroy).To(
		func(o *operation.DestroyOperation, request *operation.DestroyRequest) v1.Status {
			var err error
			if res == models.Failed {
				err = errors.New("mock error")
			}
			for _, r := range request.Intent.Resources {
				// ing -> $res
				o.MsgCh <- models.Message{
					ResourceID: r.ResourceKey(),
					OpResult:   "",
					OpErr:      nil,
				}
				o.MsgCh <- models.Message{
					ResourceID: r.ResourceKey(),
					OpResult:   res,
					OpErr:      err,
				}
			}
			close(o.MsgCh)
			if res == models.Failed {
				return v1.NewErrorStatus(err)
			}
			return nil
		}).Build()
}

func mockBackend() {
	mockey.Mock(backend.NewBackend).Return(&storages.LocalStorage{}, nil).Build()
}

func mockWorkspaceStorage() {
	mockey.Mock((*storages.LocalStorage).WorkspaceStorage).Return(&workspacestorages.LocalStorage{}, nil).Build()
	mockey.Mock((*workspacestorages.LocalStorage).GetCurrent).Return("default", nil).Build()
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

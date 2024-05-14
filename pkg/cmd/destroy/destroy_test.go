// Copyright 2024 KusionStack Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package destroy

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bytedance/mockey"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/backend/storages"
	"kusionstack.io/kusion/pkg/cmd/meta"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	releasestorages "kusionstack.io/kusion/pkg/engine/release/storages"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
	"kusionstack.io/kusion/pkg/project"
	"kusionstack.io/kusion/pkg/util/terminal"
	workspacestorages "kusionstack.io/kusion/pkg/workspace/storages"
)

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

var (
	proj = &apiv1.Project{
		Name: "fake-proj",
	}
	stack = &apiv1.Stack{
		Name: "fake-stack",
	}
	workspace = &apiv1.Workspace{
		Name: "fake-workspace",
	}
)

const (
	apiVersion = "v1"
	kind       = "ServiceAccount"
	namespace  = "test-ns"
)

var (
	sa1 = mockSA("sa1")
	sa2 = mockSA("sa2")
)

func mockSA(name string) apiv1.Resource {
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

func mockDeleteOptions() *DestroyOptions {
	return &DestroyOptions{
		MetaOptions: &meta.MetaOptions{
			RefProject:   proj,
			RefStack:     stack,
			RefWorkspace: workspace,
			Backend:      &storages.LocalStorage{},
		},
		Detail: false,
		UI:     terminal.DefaultUI(),
	}
}

func mockRelease(resources apiv1.Resources) *apiv1.Release {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return &apiv1.Release{
		Project:      "fake-proj",
		Workspace:    "fake-workspace",
		Revision:     2,
		Stack:        "fake-stack",
		Spec:         &apiv1.Spec{Resources: resources},
		State:        &apiv1.State{Resources: resources},
		Phase:        apiv1.ReleasePhaseDestroying,
		CreateTime:   time.Date(2024, 5, 21, 14, 48, 0, 0, loc),
		ModifiedTime: time.Date(2024, 5, 21, 14, 48, 0, 0, loc),
	}
}

func mockDetectProjectAndStack() {
	mockey.Mock(project.DetectProjectAndStackFrom).To(func(stackDir string) (*apiv1.Project, *apiv1.Stack, error) {
		proj.Path = stackDir
		stack.Path = stackDir
		return proj, stack, nil
	}).Build()
}

func mockNewKubernetesRuntime() {
	mockey.Mock(kubernetes.NewKubernetesRuntime).To(func() (runtime.Runtime, error) {
		return &fakerRuntime{}, nil
	}).Build()
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

func mockWorkspaceStorage() {
	mockey.Mock((*storages.LocalStorage).WorkspaceStorage).Return(&workspacestorages.LocalStorage{}, nil).Build()
	mockey.Mock((*workspacestorages.LocalStorage).GetCurrent).Return("default", nil).Build()
}

func mockReleaseStorage() {
	mockey.Mock((*storages.LocalStorage).ReleaseStorage).Return(&releasestorages.LocalStorage{}, nil).Build()
	mockey.Mock((*releasestorages.LocalStorage).Create).Return(nil).Build()
	mockey.Mock((*releasestorages.LocalStorage).Update).Return(nil).Build()
	mockey.Mock((*releasestorages.LocalStorage).GetLatestRevision).Return(1).Build()
	mockey.Mock((*releasestorages.LocalStorage).Get).Return(&apiv1.Release{State: &apiv1.State{}, Phase: apiv1.ReleasePhaseSucceeded}, nil).Build()
}

func mockPromptOutput(res string) {
	mockey.Mock((*pterm.InteractiveSelectPrinter).Show).Return(res, nil).Build()
}

func TestDestroyOptions_Run(t *testing.T) {
	mockey.PatchConvey("Detail is true", t, func() {
		mockWorkspaceStorage()
		mockNewKubernetesRuntime()
		mockOperationPreview()
		mockReleaseStorage()

		o := mockDeleteOptions()
		o.Detail = true
		err := o.Run()
		assert.Nil(t, err)
	})

	mockey.PatchConvey("prompt no", t, func() {
		mockDetectProjectAndStack()
		mockWorkspaceStorage()
		mockNewKubernetesRuntime()
		mockOperationPreview()
		mockReleaseStorage()

		o := mockDeleteOptions()
		mockPromptOutput("no")
		err := o.Run()
		assert.Nil(t, err)
	})

	mockey.PatchConvey("prompt yes", t, func() {
		mockDetectProjectAndStack()
		mockWorkspaceStorage()
		mockNewKubernetesRuntime()
		mockOperationPreview()
		mockOperationDestroy(models.Success)
		mockReleaseStorage()

		o := mockDeleteOptions()
		mockPromptOutput("yes")
		err := o.Run()
		assert.Nil(t, err)
	})
}

func TestPreview(t *testing.T) {
	mockey.PatchConvey("preview success", t, func() {
		mockNewKubernetesRuntime()
		mockOperationPreview()

		o := mockDeleteOptions()
		_, err := o.preview(&apiv1.Spec{Resources: []apiv1.Resource{sa1}}, &apiv1.State{Resources: []apiv1.Resource{sa1}}, proj, stack, &releasestorages.LocalStorage{})
		assert.Nil(t, err)
	})
}

func TestDestroy(t *testing.T) {
	mockey.PatchConvey("destroy success", t, func() {
		mockNewKubernetesRuntime()
		mockOperationDestroy(models.Success)

		o := mockDeleteOptions()
		rel := mockRelease([]apiv1.Resource{sa2})
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

		_, err := o.destroy(rel, changes, &releasestorages.LocalStorage{})
		assert.Nil(t, err)
	})
	mockey.PatchConvey("destroy failed", t, func() {
		mockNewKubernetesRuntime()
		mockOperationDestroy(models.Failed)

		o := mockDeleteOptions()
		rel := mockRelease([]apiv1.Resource{sa1})
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

		_, err := o.destroy(rel, changes, &releasestorages.LocalStorage{})
		assert.NotNil(t, err)
	})
}

func mockOperationDestroy(res models.OpResult) {
	mockey.Mock((*operation.DestroyOperation).Destroy).To(
		func(o *operation.DestroyOperation, request *operation.DestroyRequest) (*operation.DestroyResponse, v1.Status) {
			var err error
			if res == models.Failed {
				err = errors.New("mock error")
			}
			for _, r := range request.Release.State.Resources {
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
				return nil, v1.NewErrorStatus(err)
			}
			return &operation.DestroyResponse{}, nil
		}).Build()
}

func TestPrompt(t *testing.T) {
	mockey.PatchConvey("prompt error", t, func() {
		mockey.Mock((*pterm.InteractiveSelectPrinter).Show).Return("", errors.New("mock error")).Build()
		_, err := prompt(terminal.DefaultUI())
		assert.NotNil(t, err)
	})

	mockey.PatchConvey("prompt yes", t, func() {
		mockPromptOutput("yes")
		_, err := prompt(terminal.DefaultUI())
		assert.Nil(t, err)
	})
}

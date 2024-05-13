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

package apply

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/backend/storages"
	"kusionstack.io/kusion/pkg/cmd/generate"
	"kusionstack.io/kusion/pkg/cmd/meta"
	"kusionstack.io/kusion/pkg/cmd/preview"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
	statestorages "kusionstack.io/kusion/pkg/engine/state/storages"
	"kusionstack.io/kusion/pkg/util/terminal"
	workspacestorages "kusionstack.io/kusion/pkg/workspace/storages"
)

var (
	proj = &apiv1.Project{
		Name: "testdata",
	}
	stack = &apiv1.Stack{
		Name: "dev",
	}
	workspace = &apiv1.Workspace{
		Name: "default",
	}
)

func NewApplyOptions() *ApplyOptions {
	storageBackend := storages.NewLocalStorage(&apiv1.BackendLocalConfig{
		Path: filepath.Join("", "state.yaml"),
	})
	return &ApplyOptions{
		PreviewOptions: &preview.PreviewOptions{
			MetaOptions: &meta.MetaOptions{
				RefProject:     proj,
				RefStack:       stack,
				RefWorkspace:   workspace,
				StorageBackend: storageBackend,
			},
			Operator:     "",
			Detail:       false,
			All:          false,
			NoStyle:      false,
			Output:       "",
			IgnoreFields: nil,
			UI:           terminal.DefaultUI(),
		},
	}
}

func TestApplyOptionsRun(t *testing.T) {
	/*mockey.PatchConvey("Detail is true", t, func() {
		mockGenerateSpecWithSpinner()
		mockPatchNewKubernetesRuntime()
		mockPatchOperationPreview()
		mockStateStorage()

		o := NewApplyOptions()
		o.Detail = true
		o.All = true
		o.NoStyle = true
		err := o.Run()
		assert.Nil(t, err)
	})*/

	mockey.PatchConvey("DryRun is true", t, func() {
		mockGenerateSpecWithSpinner()
		mockPatchNewKubernetesRuntime()
		mockPatchOperationPreview()
		mockStateStorage()
		mockOperationApply(models.Success)

		o := NewApplyOptions()
		o.DryRun = true
		mockPromptOutput("yes")
		err := o.Run()
		assert.Nil(t, err)
	})
}

func mockGenerateSpecWithSpinner() {
	mockey.Mock(generate.GenerateSpecWithSpinner).To(func(
		project *apiv1.Project,
		stack *apiv1.Stack,
		workspace *apiv1.Workspace,
		parameters map[string]string,
		ui *terminal.UI,
		noStyle bool,
	) (*apiv1.Spec, error) {
		return &apiv1.Spec{Resources: []apiv1.Resource{sa1, sa2, sa3}}, nil
	}).Build()
}

func mockPatchNewKubernetesRuntime() *mockey.Mocker {
	return mockey.Mock(kubernetes.NewKubernetesRuntime).To(func() (runtime.Runtime, error) {
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

func mockPatchOperationPreview() *mockey.Mocker {
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

func mockStateStorage() {
	mockey.Mock((*storages.LocalStorage).WorkspaceStorage).Return(&workspacestorages.LocalStorage{}, nil).Build()
}

const (
	apiVersion = "v1"
	kind       = "ServiceAccount"
	namespace  = "test-ns"
)

var (
	sa1 = newSA("sa1")
	sa2 = newSA("sa2")
	sa3 = newSA("sa3")
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

func TestApply(t *testing.T) {
	stateStorage := statestorages.NewLocalStorage(filepath.Join("", "state.yaml"))
	mockey.PatchConvey("dry run", t, func() {
		planResources := &apiv1.Spec{Resources: []apiv1.Resource{sa1}}
		order := &models.ChangeOrder{
			StepKeys: []string{sa1.ID},
			ChangeSteps: map[string]*models.ChangeStep{
				sa1.ID: {
					ID:     sa1.ID,
					Action: models.Create,
					From:   sa1,
				},
			},
		}
		changes := models.NewChanges(proj, stack, order)
		o := NewApplyOptions()
		o.DryRun = true
		err := Apply(o, stateStorage, planResources, changes, os.Stdout)
		assert.Nil(t, err)
	})
	mockey.PatchConvey("apply success", t, func() {
		mockOperationApply(models.Success)
		o := NewApplyOptions()
		planResources := &apiv1.Spec{Resources: []apiv1.Resource{sa1, sa2}}
		order := &models.ChangeOrder{
			StepKeys: []string{sa1.ID, sa2.ID},
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
			},
		}
		changes := models.NewChanges(proj, stack, order)

		err := Apply(o, stateStorage, planResources, changes, os.Stdout)
		assert.Nil(t, err)
	})
	mockey.PatchConvey("apply failed", t, func() {
		mockOperationApply(models.Failed)

		o := NewApplyOptions()
		planResources := &apiv1.Spec{Resources: []apiv1.Resource{sa1}}
		order := &models.ChangeOrder{
			StepKeys: []string{sa1.ID},
			ChangeSteps: map[string]*models.ChangeStep{
				sa1.ID: {
					ID:     sa1.ID,
					Action: models.Create,
					From:   &sa1,
				},
			},
		}
		changes := models.NewChanges(proj, stack, order)

		err := Apply(o, stateStorage, planResources, changes, os.Stdout)
		assert.NotNil(t, err)
	})
}

func mockOperationApply(res models.OpResult) {
	mockey.Mock((*operation.ApplyOperation).Apply).To(
		func(o *operation.ApplyOperation, request *operation.ApplyRequest) (*operation.ApplyResponse, v1.Status) {
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
				return nil, v1.NewErrorStatus(err)
			}
			return &operation.ApplyResponse{}, nil
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

func mockPromptOutput(res string) {
	mockey.Mock((*pterm.InteractiveSelectPrinter).Show).Return(res, nil).Build()
}

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

package preview

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	internalv1 "kusionstack.io/kusion/pkg/apis/internal.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/backend/storages"
	"kusionstack.io/kusion/pkg/cmd/generate"
	"kusionstack.io/kusion/pkg/cmd/meta"
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
	apiVersion = "v1"
	kind       = "ServiceAccount"
	namespace  = "test-ns"

	proj = &apiv1.Project{
		Name: "testdata",
	}
	stack = &apiv1.Stack{
		Name: "dev",
	}
	workspace = &apiv1.Workspace{
		Name: "default",
	}

	sa1 = newSA("sa1")
	sa2 = newSA("sa2")
	sa3 = newSA("sa3")
)

func NewPreviewOptions() *PreviewOptions {
	storageBackend := storages.NewLocalStorage(&internalv1.BackendLocalConfig{
		Path: filepath.Join("", "state.yaml"),
	})
	return &PreviewOptions{
		MetaOptions: &meta.MetaOptions{
			RefProject:     proj,
			RefStack:       stack,
			RefWorkspace:   workspace,
			StorageBackend: storageBackend,
		},
	}
}

func TestPreview(t *testing.T) {
	stateStorage := statestorages.NewLocalStorage(filepath.Join("", "state.yaml"))
	t.Run("preview success", func(t *testing.T) {
		m := mockOperationPreview()
		defer m.UnPatch()

		o := &PreviewOptions{}
		_, err := Preview(o, stateStorage, &apiv1.Spec{Resources: []apiv1.Resource{sa1, sa2, sa3}}, proj, stack)
		assert.Nil(t, err)
	})
}

func TestPreviewOptionsRun(t *testing.T) {
	t.Run("detail is true", func(t *testing.T) {
		mockey.PatchConvey("mock engine operation", t, func() {
			mockGenerateSpecWithSpinner()
			mockNewKubernetesRuntime()
			mockOperationPreview()
			mockPromptDetail("")
			mockStateStorage()

			o := NewPreviewOptions()
			o.Detail = true
			err := o.Run()
			assert.Nil(t, err)
		})
	})

	t.Run("json output is true", func(t *testing.T) {
		mockey.PatchConvey("mock engine operation", t, func() {
			mockGenerateSpecWithSpinner()
			mockNewKubernetesRuntime()
			mockOperationPreview()
			mockPromptDetail("")
			mockStateStorage()

			o := NewPreviewOptions()
			o.Output = jsonOutput
			err := o.Run()
			assert.Nil(t, err)
		})
	})

	t.Run("no style is true", func(t *testing.T) {
		mockey.PatchConvey("mock engine operation", t, func() {
			mockGenerateSpecWithSpinner()
			mockNewKubernetesRuntime()
			mockOperationPreview()
			mockPromptDetail("")
			mockStateStorage()

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

func mockNewKubernetesRuntime() {
	mockey.Mock(kubernetes.NewKubernetesRuntime).To(func() (runtime.Runtime, error) {
		return &fooRuntime{}, nil
	}).Build()
}

func mockPromptDetail(input string) {
	mockey.Mock((*models.ChangeOrder).PromptDetails).To(func(ui *terminal.UI) (string, error) {
		return input, nil
	}).Build()
}

func mockStateStorage() {
	mockey.Mock((*storages.LocalStorage).WorkspaceStorage).Return(&workspacestorages.LocalStorage{}, nil).Build()
}

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

package api

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
	statestorages "kusionstack.io/kusion/pkg/engine/state/storages"
)

func TestDestroyPreview(t *testing.T) {
	stateStorage := statestorages.NewLocalStorage(filepath.Join("", "state.yaml"))
	mockey.PatchConvey("preview success", t, func() {
		mockNewKubernetesRuntime()
		mockOperationPreview()

		o := &APIOptions{}
		_, err := DestroyPreview(o, &apiv1.Spec{Resources: []apiv1.Resource{sa1}}, proj, stack, stateStorage)
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

func TestDestroy(t *testing.T) {
	stateStorage := statestorages.NewLocalStorage(filepath.Join("", "state.yaml"))
	mockey.PatchConvey("destroy success", t, func() {
		mockNewKubernetesRuntime()
		mockOperationDestroy(models.Success)

		o := &APIOptions{}
		planResources := &apiv1.Spec{Resources: []apiv1.Resource{sa2}}
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

		err := Destroy(o, planResources, changes, stateStorage)
		assert.Nil(t, err)
	})
	mockey.PatchConvey("destroy failed", t, func() {
		mockNewKubernetesRuntime()
		mockOperationDestroy(models.Failed)

		o := &APIOptions{}
		planResources := &apiv1.Spec{Resources: []apiv1.Resource{sa1}}
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

		err := Destroy(o, planResources, changes, stateStorage)
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

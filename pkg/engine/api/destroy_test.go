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
	"testing"
	"time"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	releasestorages "kusionstack.io/kusion/pkg/engine/release/storages"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
)

func mockDestroyRelease(resources apiv1.Resources) *apiv1.Release {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return &apiv1.Release{
		Project:      "fake-proj",
		Workspace:    "fake-workspace",
		Revision:     2,
		Stack:        "fake-stack",
		Spec:         &apiv1.Spec{Resources: resources},
		State:        &apiv1.State{Resources: resources},
		Phase:        apiv1.ReleasePhaseDestroying,
		CreateTime:   time.Date(2024, 5, 21, 15, 43, 0, 0, loc),
		ModifiedTime: time.Date(2024, 5, 21, 15, 43, 0, 0, loc),
	}
}

func TestDestroyPreview(t *testing.T) {
	mockey.PatchConvey("preview success", t, func() {
		mockNewKubernetesRuntime()
		mockOperationPreview()

		o := &APIOptions{
			MaxConcurrent: constant.MaxConcurrent,
		}

		_, err := DestroyPreview(o, &apiv1.Spec{Resources: []apiv1.Resource{sa1}}, &apiv1.State{Resources: []apiv1.Resource{sa1}}, proj, stack, &releasestorages.LocalStorage{})
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
	mockey.PatchConvey("destroy success", t, func() {
		mockNewKubernetesRuntime()
		mockOperationDestroy(models.Success)

		o := &APIOptions{
			MaxConcurrent: constant.MaxConcurrent,
		}

		rel := mockDestroyRelease([]apiv1.Resource{sa2})
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

		_, err := Destroy(o, rel, changes, &releasestorages.LocalStorage{})
		assert.Nil(t, err)
	})

	mockey.PatchConvey("destroy failed", t, func() {
		mockNewKubernetesRuntime()
		mockOperationDestroy(models.Failed)

		o := &APIOptions{
			MaxConcurrent: constant.MaxConcurrent,
		}

		rel := mockDestroyRelease([]apiv1.Resource{sa1})
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

		_, err := Destroy(o, rel, changes, &releasestorages.LocalStorage{})
		assert.NotNil(t, err)
	})
}

func mockOperationDestroy(res models.OpResult) {
	mockey.Mock((*operation.DestroyOperation).Destroy).To(
		func(o *operation.DestroyOperation, request *operation.DestroyRequest) (*operation.DestroyResponse, v1.Status) {
			st := mockOperation(res, o.MsgCh, request.Release)
			if st != nil {
				return nil, st
			}
			return &operation.DestroyResponse{}, nil
		}).Build()
}

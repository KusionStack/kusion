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
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	statestorages "kusionstack.io/kusion/pkg/engine/state/storages"
)

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
		o := &APIOptions{}
		o.DryRun = true
		err := Apply(o, stateStorage, planResources, changes, os.Stdout)
		assert.Nil(t, err)
	})
	mockey.PatchConvey("apply success", t, func() {
		mockOperationApply(models.Success)
		o := &APIOptions{}
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

		o := &APIOptions{}
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

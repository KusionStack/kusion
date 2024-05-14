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
	"testing"
	"time"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	releasestorages "kusionstack.io/kusion/pkg/engine/release/storages"
)

func mockApplyRelease(resources apiv1.Resources) *apiv1.Release {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return &apiv1.Release{
		Project:      "fake-proj",
		Workspace:    "fake-workspace",
		Revision:     2,
		Stack:        "fake-stack",
		Spec:         &apiv1.Spec{Resources: resources},
		State:        &apiv1.State{},
		Phase:        apiv1.ReleasePhaseApplying,
		CreateTime:   time.Date(2024, 5, 21, 15, 29, 0, 0, loc),
		ModifiedTime: time.Date(2024, 5, 21, 15, 29, 0, 0, loc),
	}
}

func TestApply(t *testing.T) {
	mockey.PatchConvey("dry run", t, func() {
		rel := mockApplyRelease([]apiv1.Resource{sa1})
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
		_, err := Apply(o, &releasestorages.LocalStorage{}, rel, changes, os.Stdout)
		assert.Nil(t, err)
	})
	mockey.PatchConvey("apply success", t, func() {
		mockOperationApply(models.Success)
		o := &APIOptions{}
		rel := mockApplyRelease([]apiv1.Resource{sa1, sa2})
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

		_, err := Apply(o, &releasestorages.LocalStorage{}, rel, changes, os.Stdout)
		assert.Nil(t, err)
	})
	mockey.PatchConvey("apply failed", t, func() {
		mockOperationApply(models.Failed)

		o := &APIOptions{}
		rel := mockApplyRelease([]apiv1.Resource{sa1})
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

		_, err := Apply(o, &releasestorages.LocalStorage{}, rel, changes, os.Stdout)
		assert.NotNil(t, err)
	})
}

func mockOperationApply(res models.OpResult) {
	mockey.Mock((*operation.ApplyOperation).Apply).To(
		func(o *operation.ApplyOperation, request *operation.ApplyRequest) (*operation.ApplyResponse, v1.Status) {
			st := mockOperation(res, o.MsgCh, request.Release)
			if st != nil {
				return nil, st
			}
			return &operation.ApplyResponse{}, nil
		}).Build()
}

func mockOperation(res models.OpResult, msgCh chan models.Message, rel *apiv1.Release) v1.Status {
	var err error
	if res == models.Failed {
		err = errors.New("mock error")
	}
	for _, r := range rel.State.Resources {
		// ing -> $res
		msgCh <- models.Message{
			ResourceID: r.ResourceKey(),
			OpResult:   "",
			OpErr:      nil,
		}
		msgCh <- models.Message{
			ResourceID: r.ResourceKey(),
			OpResult:   res,
			OpErr:      err,
		}
	}
	close(msgCh)
	if res == models.Failed {
		return v1.NewErrorStatus(err)
	}
	return nil
}

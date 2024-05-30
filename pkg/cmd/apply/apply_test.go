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
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/bytedance/mockey"
	"github.com/liu-hm19/pterm"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/backend/storages"
	"kusionstack.io/kusion/pkg/cmd/generate"
	"kusionstack.io/kusion/pkg/cmd/meta"
	"kusionstack.io/kusion/pkg/cmd/preview"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/printers"
	releasestorages "kusionstack.io/kusion/pkg/engine/release/storages"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
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

func newApplyOptions() *ApplyOptions {
	return &ApplyOptions{
		PreviewOptions: &preview.PreviewOptions{
			MetaOptions: &meta.MetaOptions{
				RefProject:   proj,
				RefStack:     stack,
				RefWorkspace: workspace,
				Backend:      &storages.LocalStorage{},
			},
			Detail:       false,
			All:          false,
			NoStyle:      false,
			Output:       "",
			IgnoreFields: nil,
			UI:           terminal.DefaultUI(),
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

func mockPatchNewKubernetesRuntime() *mockey.Mocker {
	return mockey.Mock(kubernetes.NewKubernetesRuntime).To(func() (runtime.Runtime, error) {
		return &fakerRuntime{}, nil
	}).Build()
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

func mockWorkspaceStorage() {
	mockey.Mock((*storages.LocalStorage).WorkspaceStorage).Return(&workspacestorages.LocalStorage{}, nil).Build()
}

func mockReleaseStorage() {
	mockey.Mock((*storages.LocalStorage).ReleaseStorage).Return(&releasestorages.LocalStorage{}, nil).Build()
	mockey.Mock((*releasestorages.LocalStorage).Create).Return(nil).Build()
	mockey.Mock((*releasestorages.LocalStorage).Update).Return(nil).Build()
	mockey.Mock((*releasestorages.LocalStorage).GetLatestRevision).Return(0).Build()
	mockey.Mock((*releasestorages.LocalStorage).Get).Return(&apiv1.Release{State: &apiv1.State{}, Phase: apiv1.ReleasePhaseSucceeded}, nil).Build()
}

func TestApplyOptions_Run(t *testing.T) {
	mockey.PatchConvey("DryRun is true", t, func() {
		mockGenerateSpecWithSpinner()
		mockPatchNewKubernetesRuntime()
		mockPatchOperationPreview()
		mockWorkspaceStorage()
		mockReleaseStorage()
		mockOperationApply(models.Success)

		o := newApplyOptions()
		o.DryRun = true
		mockPromptOutput("yes")
		err := o.Run()
		assert.Nil(t, err)
	})
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
	loc, _ := time.LoadLocation("Asia/Shanghai")
	mockey.PatchConvey("dry run", t, func() {
		mockey.Mock((*releasestorages.LocalStorage).Update).Return(nil).Build()

		rel := &apiv1.Release{
			Project:      "fake-project",
			Workspace:    "fake-workspace",
			Revision:     1,
			Stack:        "fake-stack",
			Spec:         &apiv1.Spec{Resources: []apiv1.Resource{sa1}},
			State:        &apiv1.State{},
			Phase:        apiv1.ReleasePhaseApplying,
			CreateTime:   time.Date(2024, 5, 20, 19, 39, 0, 0, loc),
			ModifiedTime: time.Date(2024, 5, 20, 19, 39, 0, 0, loc),
		}
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
		o := newApplyOptions()
		o.DryRun = true
		_, err := Apply(o, &releasestorages.LocalStorage{}, rel, changes)
		assert.Nil(t, err)
	})
	mockey.PatchConvey("apply success", t, func() {
		mockOperationApply(models.Success)
		mockey.Mock((*releasestorages.LocalStorage).Update).Return(nil).Build()

		o := newApplyOptions()
		rel := &apiv1.Release{
			Project:      "fake-project",
			Workspace:    "fake-workspace",
			Revision:     1,
			Stack:        "fake-stack",
			Spec:         &apiv1.Spec{Resources: []apiv1.Resource{sa1, sa2}},
			State:        &apiv1.State{},
			Phase:        apiv1.ReleasePhaseApplying,
			CreateTime:   time.Date(2024, 5, 20, 19, 39, 0, 0, loc),
			ModifiedTime: time.Date(2024, 5, 20, 19, 39, 0, 0, loc),
		}
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
		_, err := Apply(o, &releasestorages.LocalStorage{}, rel, changes)
		assert.Nil(t, err)
	})
	mockey.PatchConvey("apply failed", t, func() {
		mockOperationApply(models.Failed)
		mockey.Mock((*releasestorages.LocalStorage).Update).Return(nil).Build()

		o := newApplyOptions()
		rel := &apiv1.Release{
			Project:      "fake-project",
			Workspace:    "fake-workspace",
			Revision:     1,
			Stack:        "fake-stack",
			Spec:         &apiv1.Spec{Resources: []apiv1.Resource{sa1}},
			State:        &apiv1.State{},
			Phase:        apiv1.ReleasePhaseApplying,
			CreateTime:   time.Date(2024, 5, 20, 19, 39, 0, 0, loc),
			ModifiedTime: time.Date(2024, 5, 20, 19, 39, 0, 0, loc),
		}
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

		_, err := Apply(o, &releasestorages.LocalStorage{}, rel, changes)
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
			for _, r := range request.Release.Spec.Resources {
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

func mockPromptOutput(res string) {
	mockey.Mock((*pterm.InteractiveSelectPrinter).Show).Return(res, nil).Build()
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

func TestWatchK8sResources(t *testing.T) {
	t.Run("watch timeout", func(t *testing.T) {
		eventCh := make(chan watch.Event, 10)
		errCh := make(chan error, 10)

		objMap := make(map[string]interface{})
		eventCh <- watch.Event{
			Type: watch.Added,
			Object: &unstructured.Unstructured{
				Object: objMap,
			},
		}

		watchK8sResources(
			"fake-resource-id",
			[]<-chan watch.Event{
				eventCh,
			},
			&printers.Table{
				IDs: []string{
					"fake-resource-id-0",
					"fake-resource-id-1",
				},
				Rows: map[string]*printers.Row{},
			},
			map[string]*printers.Table{
				"fake-resource-id": {},
			},
			1, false, errCh)

		err := <-errCh
		assert.ErrorContains(t, err, "as timeout for")
	})
}

func TestWatchTFResources(t *testing.T) {
	t.Run("successfully apply TF resources", func(t *testing.T) {
		eventCh := make(chan runtime.TFEvent, 10)
		events := []runtime.TFEvent{
			runtime.TFApplying,
			runtime.TFApplying,
			runtime.TFSucceeded,
		}
		for _, e := range events {
			eventCh <- e
		}

		id := "hashicorp:random:random_password:example-dev-kawesome"
		table := &printers.Table{
			IDs: []string{id},
			Rows: map[string]*printers.Row{
				"hashicorp:random:random_password:example-dev-kawesome": {},
			},
		}

		watchTFResources(id, eventCh, table, true)

		assert.Equal(t, true, table.AllCompleted())
	})
}

func TestPrintTable(t *testing.T) {
	w := io.Writer(bytes.NewBufferString(""))
	id := "fake-resource-id"
	tables := map[string]*printers.Table{
		"fake-resource-id": printers.NewTable([]string{
			"fake-resource-id",
		}),
	}

	t.Run("skip unsupported resources", func(t *testing.T) {
		printTable(&w, "fake-fake-resource-id", tables)
		assert.Contains(t, w.(*bytes.Buffer).String(), "Skip monitoring unsupported resources")
	})

	t.Run("update table", func(t *testing.T) {
		printTable(&w, id, tables)
		tableStr, err := pterm.DefaultTable.
			WithStyle(pterm.NewStyle(pterm.FgDefault)).
			WithHeaderStyle(pterm.NewStyle(pterm.FgDefault)).
			WithHasHeader().WithSeparator("  ").WithData(tables[id].Print()).Srender()

		assert.Nil(t, err)
		assert.Contains(t, w.(*bytes.Buffer).String(), tableStr)
	})
}

func TestRelHandler(t *testing.T) {
	o := newApplyOptions()
	o.DryRun = true
	storage, _ = o.Backend.ReleaseStorage(o.RefProject.Name, o.RefWorkspace.Name)

	loc, _ := time.LoadLocation("Asia/Shanghai")
	rel := &apiv1.Release{
		Project:      "fake-project",
		Workspace:    "fake-workspace",
		Revision:     1,
		Stack:        "fake-stack",
		Spec:         &apiv1.Spec{Resources: []apiv1.Resource{sa1}},
		State:        &apiv1.State{},
		Phase:        apiv1.ReleasePhaseApplying,
		CreateTime:   time.Date(2024, 5, 20, 19, 39, 0, 0, loc),
		ModifiedTime: time.Date(2024, 5, 20, 19, 39, 0, 0, loc),
	}

	testcases := []struct {
		name           string
		rel            *apiv1.Release
		releaseCreated bool
		err            error
		dryRun         bool
		expectedPhase  apiv1.ReleasePhase
	}{
		{
			name:           "release applying",
			rel:            rel,
			releaseCreated: false,
			err:            nil,
			dryRun:         true,
			expectedPhase:  apiv1.ReleasePhaseApplying,
		},
		{
			name:           "release failed",
			rel:            rel,
			releaseCreated: true,
			err:            errors.New("fake error"),
			dryRun:         true,
			expectedPhase:  apiv1.ReleasePhaseFailed,
		},
		{
			name:           "release succeeded",
			rel:            rel,
			releaseCreated: true,
			err:            nil,
			dryRun:         true,
			expectedPhase:  apiv1.ReleasePhaseSucceeded,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			relHandler(tc.rel, tc.releaseCreated, tc.err, tc.dryRun)
			assert.Equal(t, tc.expectedPhase, tc.rel.Phase)
		})
	}
}

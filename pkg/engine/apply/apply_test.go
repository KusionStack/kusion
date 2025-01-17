package apply

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/liu-hm19/pterm"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/backend/storages"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/printers"
	releasestorages "kusionstack.io/kusion/pkg/engine/release/storages"
	"kusionstack.io/kusion/pkg/engine/resource/graph"
	graphstorages "kusionstack.io/kusion/pkg/engine/resource/graph/storages"
	"kusionstack.io/kusion/pkg/engine/runtime"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	applystate "kusionstack.io/kusion/pkg/engine/apply/state"
	"kusionstack.io/kusion/pkg/util/terminal"
)

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

type MockApplyOptions struct{}

func (m *MockApplyOptions) GetRefProject() *apiv1.Project     { return proj }
func (m *MockApplyOptions) GetRefStack() *apiv1.Stack         { return stack }
func (m *MockApplyOptions) GetRefWorkspace() *apiv1.Workspace { return workspace }
func (m *MockApplyOptions) GetBackend() backend.Backend       { return &storages.LocalStorage{} }
func (m *MockApplyOptions) GetNoStyle() bool                  { return true }
func (m *MockApplyOptions) GetYes() bool                      { return true }
func (m *MockApplyOptions) GetDryRun() bool                   { return false }
func (m *MockApplyOptions) GetWatch() bool                    { return false }
func (m *MockApplyOptions) GetTimeout() int                   { return 0 }
func (m *MockApplyOptions) GetPortForward() int               { return 0 }
func (m *MockApplyOptions) GetDetail() bool                   { return false }
func (m *MockApplyOptions) GetAll() bool                      { return false }
func (m *MockApplyOptions) GetOutput() string                 { return "" }
func (m *MockApplyOptions) GetSpecFile() string               { return "" }
func (m *MockApplyOptions) GetIgnoreFields() []string         { return nil }
func (m *MockApplyOptions) GetValues() []string               { return nil }
func (m *MockApplyOptions) GetUI() *terminal.UI               { return terminal.DefaultUI() }

const (
	apiVersion = "v1"
	kind       = "ServiceAccount"
	namespace  = "test-ns"
)

var (
	sa1 = newSA("sa1")
	sa2 = newSA("sa2")
)

var state = &applystate.State{
	DryRun:         true,
	Gph:            &apiv1.Graph{},
	TargetRel:      &apiv1.Release{},
	ReleaseStorage: &releasestorages.LocalStorage{},
	GraphStorage:   &graphstorages.LocalStorage{},
	RelLock:        &sync.Mutex{},
	Ls:             &applystate.LineSummary{},
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

func mockGraphStorage() {
	mockey.Mock((*storages.LocalStorage).GraphStorage).Return(&graphstorages.LocalStorage{}, nil).Build()
	mockey.Mock((*graphstorages.LocalStorage).Create).Return(nil).Build()
	mockey.Mock((*graphstorages.LocalStorage).Delete).Return(nil).Build()
	mockey.Mock((*graphstorages.LocalStorage).Update).Return(nil).Build()
	mockey.Mock((*graphstorages.LocalStorage).Get).Return(&apiv1.Graph{
		Project:   "",
		Workspace: "",
		Resources: &apiv1.GraphResources{
			WorkloadResources:   map[string]*apiv1.GraphResource{},
			DependencyResources: map[string]*apiv1.GraphResource{},
			OtherResources:      map[string]*apiv1.GraphResource{},
			ResourceIndex:       map[string]*apiv1.ResourceEntry{},
		},
	}, nil).Build()
}

func TestApply(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	o := &MockApplyOptions{}
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
		state.TargetRel = rel
		state.DryRun = true
		err := apply(o, state, changes)
		assert.Nil(t, err)
	})
	mockey.PatchConvey("apply success", t, func() {
		mockOperationApply(models.Success)
		mockey.Mock((*releasestorages.LocalStorage).Update).Return(nil).Build()
		mockey.Mock((*storages.LocalStorage).GraphStorage).Return(&graphstorages.LocalStorage{}, nil).Build()
		mockey.Mock((*graphstorages.LocalStorage).Create).Return(nil).Build()
		// mockGraphStorage()
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
		state.TargetRel = rel
		state.Gph = &apiv1.Graph{
			Project:   rel.Project,
			Workspace: rel.Workspace,
		}
		state.DryRun = false
		graph.GenerateGraph(rel.Spec.Resources, state.Gph)
		err := apply(o, state, changes)
		assert.Nil(t, err)
	})
	mockey.PatchConvey("apply failed", t, func() {
		mockOperationApply(models.Failed)
		mockey.Mock((*releasestorages.LocalStorage).Update).Return(nil).Build()
		mockGraphStorage()
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
		state.Gph = &apiv1.Graph{}
		state.TargetRel = rel
		state.DryRun = false
		graph.GenerateGraph(rel.Spec.Resources, state.Gph)
		err := apply(o, state, changes)
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
		_, err := prompt(terminal.DefaultUI(), &applystate.State{})
		assert.NotNil(t, err)
	})

	mockey.PatchConvey("prompt yes", t, func() {
		mockPromptOutput("yes")
		_, err := prompt(terminal.DefaultUI(), &applystate.State{})
		assert.Nil(t, err)
	})
}

func TestWatchK8sResources(t *testing.T) {
	t.Run("successfully apply default K8s resources", func(t *testing.T) {
		id := "v1:Namespace:example"
		chs := make([]<-chan watch.Event, 1)
		events := []watch.Event{
			{
				Type: watch.Added,
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Namespace",
						"metadata": map[string]interface{}{
							"name": "example",
						},
						"spec": map[string]interface{}{},
					},
				},
			},
			{
				Type: watch.Added,
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Namespace",
						"metadata": map[string]interface{}{
							"name": "example",
						},
						"spec": map[string]interface{}{},
						"status": map[string]interface{}{
							"phase": corev1.NamespaceActive,
						},
					},
				},
			},
		}

		out := make(chan watch.Event, 10)
		for _, e := range events {
			out <- e
		}
		chs[0] = out
		table := &printers.Table{
			IDs:  []string{id},
			Rows: map[string]*printers.Row{},
		}
		tables := map[string]*printers.Table{
			id: table,
		}
		resource := &apiv1.GraphResource{
			ID:              id,
			Type:            "",
			Name:            "",
			CloudResourceID: "",
			Status:          "",
			Dependents:      []string{},
			Dependencies:    []string{},
		}
		gph := &apiv1.Graph{
			Project:   "example project",
			Workspace: "example workspace",
			Resources: &apiv1.GraphResources{
				WorkloadResources:   map[string]*apiv1.GraphResource{"id": resource},
				DependencyResources: map[string]*apiv1.GraphResource{},
				OtherResources:      map[string]*apiv1.GraphResource{},
				ResourceIndex:       map[string]*apiv1.ResourceEntry{},
			},
		}

		healthPolicy := map[string]interface{}{
			"health.kcl": "assert res.metadata.generation == 1",
		}

		graph.UpdateResourceIndex(gph.Resources)
		watchResult := make(chan error)
		watchK8sResources(id, "", chs, table, tables, healthPolicy, state.Gph.Resources, watchResult)

		assert.Equal(t, true, table.AllCompleted())
	})
	t.Run("successfully apply customized K8s resources", func(t *testing.T) {
		id := "v1:Deployment:example"
		chs := make([]<-chan watch.Event, 1)
		events := []watch.Event{
			{
				Type: watch.Added,
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Deployment",
						"metadata": map[string]interface{}{
							"name":       "example",
							"generation": 1,
						},
						"spec": map[string]interface{}{},
					},
				},
			},
		}

		out := make(chan watch.Event, 10)
		for _, e := range events {
			out <- e
		}
		chs[0] = out
		table := &printers.Table{
			IDs:  []string{id},
			Rows: map[string]*printers.Row{},
		}
		tables := map[string]*printers.Table{
			id: table,
		}
		var policyInterface interface{}
		healthPolicy := map[string]interface{}{
			"health.kcl": "assert res.metadata.generation == 1",
		}
		policyInterface = healthPolicy
		resource := &apiv1.GraphResource{
			ID:              id,
			Type:            "",
			Name:            "",
			CloudResourceID: "",
			Status:          "",
			Dependents:      []string{},
			Dependencies:    []string{},
		}
		gph := &apiv1.Graph{
			Project:   "example project",
			Workspace: "example workspace",
			Resources: &apiv1.GraphResources{
				WorkloadResources:   map[string]*apiv1.GraphResource{"id": resource},
				DependencyResources: map[string]*apiv1.GraphResource{},
				OtherResources:      map[string]*apiv1.GraphResource{},
				ResourceIndex:       map[string]*apiv1.ResourceEntry{},
			},
		}
		graph.UpdateResourceIndex(gph.Resources)
		watchResult := make(chan error)
		watchK8sResources(id, "Deployment", chs, table, tables, policyInterface, state.Gph.Resources, watchResult)

		assert.Equal(t, true, table.AllCompleted())
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

		watchResult := make(chan error)
		watchTFResources(id, eventCh, table, watchResult)

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

func TestGetResourceInfo(t *testing.T) {
	tests := []struct {
		name         string
		resource     *apiv1.Resource
		expectedKind string
		expectPanic  bool
	}{
		{
			name: "with valid resource",
			resource: &apiv1.Resource{
				Attributes: map[string]interface{}{
					apiv1.FieldKind: "Service",
				},
				Extensions: map[string]interface{}{
					apiv1.FieldHealthPolicy: "policyValue",
				},
			},
			expectedKind: "Service",
			expectPanic:  false,
		},
		{
			name: "with nil Attributes",
			resource: &apiv1.Resource{
				Attributes: nil,
				Extensions: map[string]interface{}{
					apiv1.FieldHealthPolicy: "policyValue",
				},
			},
			expectPanic: true,
		},
		{
			name: "with non-string kind",
			resource: &apiv1.Resource{
				Attributes: map[string]interface{}{
					apiv1.FieldKind: 123,
				},
				Extensions: map[string]interface{}{
					apiv1.FieldHealthPolicy: "policyValue",
				},
			},
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected panic for test case '%s', but got none", tt.name)
					}
				}()
			}

			healthPolicy, kind := getResourceInfo(tt.resource)
			if !tt.expectPanic {
				if kind != tt.expectedKind {
					t.Errorf("expected kind '%s', but got '%s'", tt.expectedKind, kind)
				}
				if healthPolicy != "policyValue" && !tt.expectPanic {
					t.Errorf("expected healthPolicy to be 'policyValue', but got '%v'", healthPolicy)
				}
			}
		})
	}
}

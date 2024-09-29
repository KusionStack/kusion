package operation

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation/graph"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/engine/release/storages"
	resourcegraph "kusionstack.io/kusion/pkg/engine/resource/graph"
	"kusionstack.io/kusion/pkg/engine/runtime"
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
	"kusionstack.io/kusion/third_party/terraform/dag"
)

func TestApplyOperation_Apply(t *testing.T) {
	type fields struct {
		operationType           models.OperationType
		releaseStorage          release.Storage
		ctxResourceIndex        map[string]*apiv1.Resource
		priorStateResourceIndex map[string]*apiv1.Resource
		stateResourceIndex      map[string]*apiv1.Resource
		order                   *models.ChangeOrder
		runtimeMap              map[apiv1.Type]runtime.Runtime
		stack                   *apiv1.Stack
		msgCh                   chan models.Message
		release                 *apiv1.Release
		lock                    *sync.Mutex
	}
	type args struct {
		applyRequest *ApplyRequest
	}

	fakeSpec := &apiv1.Spec{
		Resources: []apiv1.Resource{
			{
				ID:   "mock-id",
				Type: runtime.Kubernetes,
				Attributes: map[string]interface{}{
					"a": "b",
				},
				DependsOn: nil,
			},
		},
	}
	fakeState := &apiv1.State{
		Resources: []apiv1.Resource{
			{
				ID:   "mock-id",
				Type: runtime.Kubernetes,
				Attributes: map[string]interface{}{
					"a": "b",
				},
				DependsOn: nil,
			},
		},
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")
	fakeTime := time.Date(2024, 5, 10, 16, 48, 0, 0, loc)
	fakeRelease := &apiv1.Release{
		Project:      "fake-project",
		Workspace:    "fake-workspace",
		Revision:     1,
		Stack:        "fake-stack",
		Spec:         fakeSpec,
		State:        &apiv1.State{},
		Phase:        apiv1.ReleasePhaseApplying,
		CreateTime:   fakeTime,
		ModifiedTime: fakeTime,
	}
	fakeUpdateRelease := &apiv1.Release{
		Project:      "fake-project",
		Workspace:    "fake-workspace",
		Revision:     1,
		Stack:        "fake-stack",
		Spec:         fakeSpec,
		State:        fakeState,
		Phase:        apiv1.ReleasePhaseApplying,
		CreateTime:   fakeTime,
		ModifiedTime: fakeTime,
	}

	fakeStack := &apiv1.Stack{
		Name: "fake-stack",
		Path: "fake-path",
	}
	fakeProject := &apiv1.Project{
		Name:   "fake-project",
		Path:   "fake-path",
		Stacks: []*apiv1.Stack{fakeStack},
	}
	fakeGraph := &apiv1.Graph{Project: fakeRelease.Project, Workspace: fakeRelease.Workspace}
	resourcegraph.GenerateGraph(fakeSpec.Resources, fakeGraph)
	testcases := []struct {
		name             string
		fields           fields
		args             args
		expectedResponse *ApplyResponse
		expectedStatus   v1.Status
	}{
		{
			name: "apply test",
			fields: fields{
				operationType:  models.Apply,
				releaseStorage: &storages.LocalStorage{},
				runtimeMap:     map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: &kubernetes.KubernetesRuntime{}},
				msgCh:          make(chan models.Message, 5),
			},
			args: args{applyRequest: &ApplyRequest{
				Request: models.Request{
					Stack:   fakeStack,
					Project: fakeProject,
				},
				Release: fakeRelease,
				Graph:   fakeGraph,
			}},
			expectedResponse: &ApplyResponse{Release: fakeUpdateRelease, Graph: fakeGraph},
			expectedStatus:   nil,
		},
	}

	for _, tc := range testcases {
		mockey.PatchConvey(tc.name, t, func() {
			o := &models.Operation{
				OperationType:           tc.fields.operationType,
				ReleaseStorage:          tc.fields.releaseStorage,
				CtxResourceIndex:        tc.fields.ctxResourceIndex,
				PriorStateResourceIndex: tc.fields.priorStateResourceIndex,
				StateResourceIndex:      tc.fields.stateResourceIndex,
				ChangeOrder:             tc.fields.order,
				RuntimeMap:              tc.fields.runtimeMap,
				Stack:                   tc.fields.stack,
				MsgCh:                   tc.fields.msgCh,
				Release:                 tc.fields.release,
				Lock:                    tc.fields.lock,
			}
			ao := &ApplyOperation{
				Operation: *o,
			}

			mockey.Mock((*graph.ResourceNode).Execute).To(func(operation *models.Operation) v1.Status {
				operation.Release = fakeUpdateRelease
				return nil
			}).Build()
			mockey.Mock(runtimeinit.Runtimes).To(func(
				spec apiv1.Spec, state apiv1.State,
			) (map[apiv1.Type]runtime.Runtime, v1.Status) {
				return map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: &kubernetes.KubernetesRuntime{}}, nil
			}).Build()
			mockey.Mock(populateResourceGraph).Return(fakeGraph).Build()
			rsp, status := ao.Apply(tc.args.applyRequest)
			assert.Equal(t, tc.expectedResponse, rsp)
			assert.Equal(t, tc.expectedStatus, status)
		})
	}
}

func Test_ValidateApplyRequest(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		req     *ApplyRequest
	}{
		{
			name:    "invalid request nil request",
			success: false,
			req:     nil,
		},
		{
			name:    "invalid request invalid release phase",
			success: false,
			req: &ApplyRequest{
				Release: &apiv1.Release{
					Phase: "invalid_phase",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock valid release and spec", t, func() {
				mockey.Mock(release.ValidateRelease).Return(nil).Build()
				mockey.Mock(release.ValidateSpec).Return(nil).Build()
				mockey.Mock(resourcegraph.ValidateGraph).Return(nil).Build()
				err := validateApplyRequest(tc.req)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func Test_populateResourceGraph(t *testing.T) {
	graph := &dag.AcyclicGraph{
		Graph: dag.Graph{},
	}
	graph.Add("mock-ID1")
	graph.Add("mock-ID2")
	graph.Add("mock-ID")
	graph.Connect(dag.BasicEdge("mock-ID", "mock-ID1"))
	graph.Connect(dag.BasicEdge("mock-ID2", "mock-ID"))
	testResource := &apiv1.GraphResource{
		ID:              "mock-ID",
		Type:            "mock-type",
		Name:            "mock-name",
		CloudResourceID: "",
		Status:          "",
		Dependents:      []string{},
		Dependencies:    []string{},
	}
	mockResource := &apiv1.GraphResource{
		ID:              "mock-ID",
		Type:            "mock-type",
		Name:            "mock-name",
		CloudResourceID: "",
		Status:          "",
		Dependents:      []string{"mock-ID1"},
		Dependencies:    []string{"mock-ID2"},
	}
	type args struct {
		applyGraph    *dag.AcyclicGraph
		resourceGraph *apiv1.Graph
	}
	tests := []struct {
		name string
		args args
		want *apiv1.Graph
	}{
		{
			name: "poplute resource dependents and dependencies",
			args: args{
				applyGraph: graph,
				resourceGraph: &apiv1.Graph{
					Project:   "project name",
					Workspace: "workspace name",
					Resources: &apiv1.GraphResources{
						WorkloadResources:   map[string]*apiv1.GraphResource{"mock-ID": testResource},
						DependencyResources: map[string]*apiv1.GraphResource{},
						OtherResources:      map[string]*apiv1.GraphResource{},
						ResourceIndex:       map[string]*apiv1.ResourceEntry{},
					},
				},
			},
			want: &apiv1.Graph{
				Project:   "project name",
				Workspace: "workspace name",
				Resources: &apiv1.GraphResources{
					WorkloadResources:   map[string]*apiv1.GraphResource{"mock-ID": mockResource},
					DependencyResources: map[string]*apiv1.GraphResource{},
					OtherResources:      map[string]*apiv1.GraphResource{},
					ResourceIndex:       map[string]*apiv1.ResourceEntry{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourcegraph.UpdateResourceIndex(tt.args.resourceGraph.Resources)
			got := populateResourceGraph(tt.args.applyGraph, tt.args.resourceGraph)
			if !reflect.DeepEqual(got.Resources.WorkloadResources["mock-ID"].Dependents, tt.want.Resources.WorkloadResources["mock-ID"].Dependents) {
				t.Errorf("populateResourceGraph() = %v, want %v", got.Resources.WorkloadResources["mock-ID"], tt.want.Resources.WorkloadResources["mock-ID"])
			}
			if !reflect.DeepEqual(got.Resources.WorkloadResources["mock-ID"].Dependencies, tt.want.Resources.WorkloadResources["mock-ID"].Dependencies) {
				t.Errorf("populateResourceGraph() = %v, want %v", got.Resources.WorkloadResources["mock-ID"], tt.want.Resources.WorkloadResources["mock-ID"])
			}
		})
	}
}

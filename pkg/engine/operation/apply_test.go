//go:build !arm64
// +build !arm64

package operation

import (
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/bytedance/mockey"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/engine/operation/graph"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/engine/states/local"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/status"
)

func Test_validateRequest(t *testing.T) {
	type args struct {
		request *opsmodels.Request
	}
	tests := []struct {
		name string
		args args
		want status.Status
	}{
		{
			name: "t1",
			args: args{
				request: &opsmodels.Request{},
			},
			want: status.NewErrorStatusWithMsg(status.InvalidArgument,
				"request.Intent is empty. If you want to delete all resources, please use command 'destroy'"),
		},
		{
			name: "t2",
			args: args{
				request: &opsmodels.Request{
					Spec: &models.Intent{Resources: []models.Resource{}},
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateRequest(tt.args.request); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("validateRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperation_Apply(t *testing.T) {
	type fields struct {
		OperationType           opsmodels.OperationType
		StateStorage            states.StateStorage
		CtxResourceIndex        map[string]*models.Resource
		PriorStateResourceIndex map[string]*models.Resource
		StateResourceIndex      map[string]*models.Resource
		Order                   *opsmodels.ChangeOrder
		RuntimeMap              map[models.Type]runtime.Runtime
		Stack                   *projectstack.Stack
		MsgCh                   chan opsmodels.Message
		resultState             *states.State
		lock                    *sync.Mutex
	}
	type args struct {
		applyRequest *ApplyRequest
	}

	const Jack = "jack"
	mf := &models.Intent{Resources: []models.Resource{
		{
			ID:   Jack,
			Type: runtime.Kubernetes,
			Attributes: map[string]interface{}{
				"a": "b",
			},
			DependsOn: nil,
		},
	}}

	rs := &states.State{
		ID:            0,
		Tenant:        "fakeTenant",
		Stack:         "fakeStack",
		Project:       "fakeProject",
		Version:       0,
		KusionVersion: "",
		Serial:        1,
		Operator:      "faker",
		Resources: []models.Resource{
			{
				ID:   Jack,
				Type: runtime.Kubernetes,
				Attributes: map[string]interface{}{
					"a": "b",
				},
				DependsOn: nil,
			},
		},
	}

	stack := &projectstack.Stack{
		StackConfiguration: projectstack.StackConfiguration{Name: "fakeStack"},
		Path:               "fakePath",
	}
	project := &projectstack.Project{
		ProjectConfiguration: projectstack.ProjectConfiguration{
			Name:    "fakeProject",
			Tenant:  "fakeTenant",
			Backend: nil,
		},
		Path:   "fakePath",
		Stacks: []*projectstack.Stack{stack},
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRsp *ApplyResponse
		wantSt  status.Status
	}{
		{
			name: "apply test",
			fields: fields{
				OperationType: opsmodels.Apply,
				StateStorage:  &local.FileSystemState{Path: filepath.Join("test_data", local.KusionState)},
				RuntimeMap:    map[models.Type]runtime.Runtime{runtime.Kubernetes: &kubernetes.KubernetesRuntime{}},
				MsgCh:         make(chan opsmodels.Message, 5),
			},
			args: args{applyRequest: &ApplyRequest{opsmodels.Request{
				Tenant:   "fakeTenant",
				Stack:    stack,
				Project:  project,
				Operator: "faker",
				Spec:     mf,
			}}},
			wantRsp: &ApplyResponse{rs},
			wantSt:  nil,
		},
	}

	for _, tt := range tests {
		mockey.PatchConvey(tt.name, t, func() {
			o := &opsmodels.Operation{
				OperationType:           tt.fields.OperationType,
				StateStorage:            tt.fields.StateStorage,
				CtxResourceIndex:        tt.fields.CtxResourceIndex,
				PriorStateResourceIndex: tt.fields.PriorStateResourceIndex,
				StateResourceIndex:      tt.fields.StateResourceIndex,
				ChangeOrder:             tt.fields.Order,
				RuntimeMap:              tt.fields.RuntimeMap,
				Stack:                   tt.fields.Stack,
				MsgCh:                   tt.fields.MsgCh,
				ResultState:             tt.fields.resultState,
				Lock:                    tt.fields.lock,
			}
			ao := &ApplyOperation{
				Operation: *o,
			}

			mockey.Mock((*graph.ResourceNode).Execute).To(func(rn *graph.ResourceNode, operation *opsmodels.Operation) status.Status {
				o.ResultState = rs
				return nil
			}).Build()
			mockey.Mock(runtimeinit.Runtimes).To(func(
				resources models.Resources,
				stack *projectstack.Stack,
			) (map[models.Type]runtime.Runtime, status.Status) {
				return map[models.Type]runtime.Runtime{runtime.Kubernetes: &kubernetes.KubernetesRuntime{}}, nil
			}).Build()

			gotRsp, gotSt := ao.Apply(tt.args.applyRequest)
			assert.Equalf(t, tt.wantRsp.State.Stack, gotRsp.State.Stack, "Apply(%v)", tt.args.applyRequest)
			assert.Equalf(t, tt.wantSt, gotSt, "Apply(%v)", tt.args.applyRequest)
		})
	}
}

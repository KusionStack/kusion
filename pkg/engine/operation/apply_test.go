package operation

import (
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/bytedance/mockey"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation/graph"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
	"kusionstack.io/kusion/pkg/engine/state"
	"kusionstack.io/kusion/pkg/engine/state/storages"
)

func Test_ValidateRequest(t *testing.T) {
	type args struct {
		request *models.Request
	}
	tests := []struct {
		name string
		args args
		want v1.Status
	}{
		{
			name: "t1",
			args: args{
				request: &models.Request{},
			},
			want: v1.NewErrorStatusWithMsg(v1.InvalidArgument,
				"request.Intent is empty. If you want to delete all resources, please use command 'destroy'"),
		},
		{
			name: "t2",
			args: args{
				request: &models.Request{
					Intent: &apiv1.Spec{Resources: []apiv1.Resource{}},
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
		OperationType           models.OperationType
		StateStorage            state.Storage
		CtxResourceIndex        map[string]*apiv1.Resource
		PriorStateResourceIndex map[string]*apiv1.Resource
		StateResourceIndex      map[string]*apiv1.Resource
		Order                   *models.ChangeOrder
		RuntimeMap              map[apiv1.Type]runtime.Runtime
		Stack                   *apiv1.Stack
		MsgCh                   chan models.Message
		resultState             *apiv1.DeprecatedState
		lock                    *sync.Mutex
	}
	type args struct {
		applyRequest *ApplyRequest
	}

	const Jack = "jack"
	mf := &apiv1.Spec{Resources: []apiv1.Resource{
		{
			ID:   Jack,
			Type: runtime.Kubernetes,
			Attributes: map[string]interface{}{
				"a": "b",
			},
			DependsOn: nil,
		},
	}}

	rs := &apiv1.DeprecatedState{
		ID:            0,
		Stack:         "fake-stack",
		Project:       "fake-project",
		Workspace:     "fake-workspace",
		Version:       0,
		KusionVersion: "",
		Serial:        1,
		Operator:      "fake-operator",
		Resources: []apiv1.Resource{
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

	s := &apiv1.Stack{
		Name: "fake-stack",
		Path: "fake-path",
	}
	p := &apiv1.Project{
		Name:   "fake-project",
		Path:   "fake-path",
		Stacks: []*apiv1.Stack{s},
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRsp *ApplyResponse
		wantSt  v1.Status
	}{
		{
			name: "apply test",
			fields: fields{
				OperationType: models.Apply,
				StateStorage:  storages.NewLocalStorage(filepath.Join("testdata", "state.yaml")),
				RuntimeMap:    map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: &kubernetes.KubernetesRuntime{}},
				MsgCh:         make(chan models.Message, 5),
			},
			args: args{applyRequest: &ApplyRequest{models.Request{
				Stack:    s,
				Project:  p,
				Operator: "fake-operator",
				Intent:   mf,
			}}},
			wantRsp: &ApplyResponse{rs},
			wantSt:  nil,
		},
	}

	for _, tt := range tests {
		mockey.PatchConvey(tt.name, t, func() {
			o := &models.Operation{
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

			mockey.Mock((*graph.ResourceNode).Execute).To(func(operation *models.Operation) v1.Status {
				o.ResultState = rs
				return nil
			}).Build()
			mockey.Mock(runtimeinit.Runtimes).To(func(
				resources apiv1.Resources,
			) (map[apiv1.Type]runtime.Runtime, v1.Status) {
				return map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: &kubernetes.KubernetesRuntime{}}, nil
			}).Build()

			gotRsp, gotSt := ao.Apply(tt.args.applyRequest)
			assert.Equalf(t, tt.wantRsp.State.Stack, gotRsp.State.Stack, "Apply(%v)", tt.args.applyRequest)
			assert.Equalf(t, tt.wantSt, gotSt, "Apply(%v)", tt.args.applyRequest)
		})
	}
}

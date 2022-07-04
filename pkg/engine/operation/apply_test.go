//go:build !arm64
// +build !arm64

package operation

import (
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"kusionstack.io/kusion/pkg/engine/states/local"

	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"

	"kusionstack.io/kusion/pkg/engine/operation/graph"
	"kusionstack.io/kusion/pkg/engine/operation/types"

	"bou.ke/monkey"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/states"
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
				"request.Spec is empty. If you want to delete all resources, please use command 'destroy'"),
		},
		{
			name: "t2",
			args: args{
				request: &opsmodels.Request{
					Spec: &models.Spec{Resources: []models.Resource{}},
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
		OperationType           types.OperationType
		StateStorage            states.StateStorage
		CtxResourceIndex        map[string]*models.Resource
		PriorStateResourceIndex map[string]*models.Resource
		StateResourceIndex      map[string]*models.Resource
		Order                   *opsmodels.ChangeOrder
		Runtime                 runtime.Runtime
		MsgCh                   chan opsmodels.Message
		resultState             *states.State
		lock                    *sync.Mutex
	}
	type args struct {
		applyRequest *ApplyRequest
	}

	const Jack = "jack"
	mf := &models.Spec{Resources: []models.Resource{
		{
			ID: Jack,
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
				ID: Jack,
				Attributes: map[string]interface{}{
					"a": "b",
				},
				DependsOn: nil,
			},
		},
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
				OperationType: types.Apply,
				StateStorage:  &local.FileSystemState{Path: filepath.Join("test_data", local.KusionState)},
				Runtime:       &runtime.KubernetesRuntime{},
				MsgCh:         make(chan opsmodels.Message, 5),
			},
			args: args{applyRequest: &ApplyRequest{opsmodels.Request{
				Tenant:   "fakeTenant",
				Stack:    "fakeStack",
				Project:  "fakeProject",
				Operator: "faker",
				Spec:     mf,
			}}},
			wantRsp: &ApplyResponse{rs},
			wantSt:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &opsmodels.Operation{
				OperationType:           tt.fields.OperationType,
				StateStorage:            tt.fields.StateStorage,
				CtxResourceIndex:        tt.fields.CtxResourceIndex,
				PriorStateResourceIndex: tt.fields.PriorStateResourceIndex,
				StateResourceIndex:      tt.fields.StateResourceIndex,
				ChangeOrder:             tt.fields.Order,
				Runtime:                 tt.fields.Runtime,
				MsgCh:                   tt.fields.MsgCh,
				ResultState:             tt.fields.resultState,
				Lock:                    tt.fields.lock,
			}
			ao := &ApplyOperation{
				Operation: *o,
			}

			monkey.Patch((*graph.ResourceNode).Execute, func(rn *graph.ResourceNode, operation *opsmodels.Operation) status.Status {
				o.ResultState = rs
				return nil
			})

			gotRsp, gotSt := ao.Apply(tt.args.applyRequest)
			assert.Equalf(t, tt.wantRsp.State.Stack, gotRsp.State.Stack, "Apply(%v)", tt.args.applyRequest)
			assert.Equalf(t, tt.wantSt, gotSt, "Apply(%v)", tt.args.applyRequest)
		})
	}
}

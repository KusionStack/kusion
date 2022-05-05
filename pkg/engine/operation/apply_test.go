package operation

import (
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"bou.ke/monkey"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/engine/manifest"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/status"
)

func Test_validateRequest(t *testing.T) {
	type args struct {
		request *Request
	}
	tests := []struct {
		name string
		args args
		want status.Status
	}{
		{
			name: "t1",
			args: args{
				request: &Request{},
			},
			want: status.NewErrorStatusWithMsg(status.InvalidArgument,
				"request.Manifest is empty. If you want to delete all resources, please use command 'destroy'"),
		},
		{
			name: "t2",
			args: args{
				request: &Request{
					Manifest: &manifest.Manifest{Resources: []states.ResourceState{}},
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
		OperationType           Type
		StateStorage            states.StateStorage
		CtxResourceIndex        map[string]*states.ResourceState
		PriorStateInstanceIndex map[string]*states.ResourceState
		StateResourceIndex      map[string]*states.ResourceState
		ChangeStepMap           map[string]*ChangeStep
		Runtime                 runtime.Runtime
		MsgCh                   chan Message
		resultState             *states.State
		lock                    *sync.Mutex
	}
	type args struct {
		applyRequest *ApplyRequest
	}

	const Jack = "jack"
	mf := &manifest.Manifest{Resources: []states.ResourceState{
		{
			ID:   Jack,
			Mode: states.Managed,
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
		Resources: []states.ResourceState{
			{
				ID:   Jack,
				Mode: states.Managed,
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
				OperationType: Apply,
				StateStorage:  &states.FileSystemState{Path: filepath.Join("test_data", states.KusionState)},
				Runtime:       &runtime.KubernetesRuntime{},
				MsgCh:         make(chan Message, 5),
			},
			args: args{applyRequest: &ApplyRequest{Request{
				Tenant:   "fakeTenant",
				Stack:    "fakeStack",
				Project:  "fakeProject",
				Operator: "faker",
				Manifest: mf,
			}}},
			wantRsp: &ApplyResponse{rs},
			wantSt:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Operation{
				OperationType:           tt.fields.OperationType,
				StateStorage:            tt.fields.StateStorage,
				CtxResourceIndex:        tt.fields.CtxResourceIndex,
				PriorStateResourceIndex: tt.fields.PriorStateInstanceIndex,
				StateResourceIndex:      tt.fields.StateResourceIndex,
				ChangeStepMap:           tt.fields.ChangeStepMap,
				Runtime:                 tt.fields.Runtime,
				MsgCh:                   tt.fields.MsgCh,
				resultState:             tt.fields.resultState,
				lock:                    tt.fields.lock,
			}

			monkey.Patch((*ResourceNode).Execute, func(rn *ResourceNode, operation Operation) status.Status {
				o.resultState = rs
				return nil
			})

			gotRsp, gotSt := o.Apply(tt.args.applyRequest)
			assert.Equalf(t, tt.wantRsp.State.Stack, gotRsp.State.Stack, "Apply(%v)", tt.args.applyRequest)
			assert.Equalf(t, tt.wantSt, gotSt, "Apply(%v)", tt.args.applyRequest)
		})
	}
}

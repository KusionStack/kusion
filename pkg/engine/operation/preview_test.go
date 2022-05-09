package operation

import (
	"reflect"
	"sync"
	"testing"

	"kusionstack.io/kusion/pkg/engine/manifest"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/status"
	"kusionstack.io/kusion/pkg/util/kdump"
)

var (
	FakeService = map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata": map[string]interface{}{
			"name":      "apple-service",
			"namespace": "http-echo",
		},
		"spec": map[string]interface{}{
			"type": "NodePort",
		},
	}
	FakeResourceState = states.ResourceState{
		ID:         "fake-id",
		Mode:       states.Managed,
		Attributes: FakeService,
	}
)

func TestOperation_Preview(t *testing.T) {
	type fields struct {
		OperationType           Type
		StateStorage            states.StateStorage
		CtxResourceIndex        map[string]*states.ResourceState
		PriorStateResourceIndex map[string]*states.ResourceState
		StateResourceIndex      map[string]*states.ResourceState
		ChangeStepMap           map[string]*ChangeStep
		Runtime                 runtime.Runtime
		MsgCh                   chan Message
		resultState             *states.State
		lock                    *sync.Mutex
	}
	type args struct {
		request   *PreviewRequest
		operation Type
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRsp *PreviewResponse
		wantS   status.Status
	}{
		{
			name: "success-when-apply",
			fields: fields{
				Runtime:       &runtime.KubernetesRuntime{},
				StateStorage:  &states.FileSystemState{Path: states.KusionState},
				ChangeStepMap: map[string]*ChangeStep{},
			},
			args: args{
				request: &PreviewRequest{
					Request: Request{
						Tenant:   "fake-tennat",
						Stack:    "fake-stack",
						Project:  "fake-project",
						Operator: "fake-operator",
						Manifest: &manifest.Manifest{
							Resources: []states.ResourceState{
								FakeResourceState,
							},
						},
					},
				},
				operation: Apply,
			},
			wantRsp: &PreviewResponse{
				ChangeSteps: map[string]*ChangeStep{
					"fake-id": {
						ID:     "fake-id",
						Action: Create,
						Old:    (*states.ResourceState)(nil),
						New:    &FakeResourceState,
					},
				},
			},
			wantS: nil,
		},
		{
			name: "success-when-destroy",
			fields: fields{
				Runtime:       &runtime.KubernetesRuntime{},
				StateStorage:  &states.FileSystemState{Path: states.KusionState},
				ChangeStepMap: map[string]*ChangeStep{},
			},
			args: args{
				request: &PreviewRequest{
					Request: Request{
						Tenant:   "fake-tennat",
						Stack:    "fake-stack",
						Project:  "fake-project",
						Operator: "fake-operator",
						Manifest: &manifest.Manifest{
							Resources: []states.ResourceState{
								FakeResourceState,
							},
						},
					},
				},
				operation: Destroy,
			},
			wantRsp: &PreviewResponse{
				ChangeSteps: map[string]*ChangeStep{
					"fake-id": {
						ID:     "fake-id",
						Action: Delete,
						Old:    &FakeResourceState,
						New:    (*states.ResourceState)(nil),
					},
				},
			},
			wantS: nil,
		},
		{
			name: "fail-because-empty-manifest",
			fields: fields{
				Runtime:       &runtime.KubernetesRuntime{},
				StateStorage:  &states.FileSystemState{Path: states.KusionState},
				ChangeStepMap: map[string]*ChangeStep{},
			},
			args: args{
				request: &PreviewRequest{
					Request: Request{
						Manifest: nil,
					},
				},
				operation: Apply,
			},
			wantRsp: nil,
			wantS:   status.NewErrorStatusWithMsg(status.InvalidArgument, "request.Manifest is empty. If you want to delete all resources, please use command 'destroy'"),
		},
		{
			name: "fail-because-nonexistent-id",
			fields: fields{
				Runtime:       &runtime.KubernetesRuntime{},
				StateStorage:  &states.FileSystemState{Path: states.KusionState},
				ChangeStepMap: map[string]*ChangeStep{},
			},
			args: args{
				request: &PreviewRequest{
					Request: Request{
						Tenant:   "fake-tennat",
						Stack:    "fake-stack",
						Project:  "fake-project",
						Operator: "fake-operator",
						Manifest: &manifest.Manifest{
							Resources: []states.ResourceState{
								{
									ID:         "fake-id",
									Mode:       states.Managed,
									Attributes: FakeService,
									DependsOn:  []string{"nonexistent-id"},
								},
							},
						},
					},
				},
				operation: Apply,
			},
			wantRsp: nil,
			wantS:   status.NewErrorStatusWithMsg(status.IllegalManifest, "can't find resource by key:nonexistent-id in manifest or state."),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Operation{
				OperationType:           tt.fields.OperationType,
				StateStorage:            tt.fields.StateStorage,
				CtxResourceIndex:        tt.fields.CtxResourceIndex,
				PriorStateResourceIndex: tt.fields.PriorStateResourceIndex,
				StateResourceIndex:      tt.fields.StateResourceIndex,
				ChangeStepMap:           tt.fields.ChangeStepMap,
				Runtime:                 tt.fields.Runtime,
				MsgCh:                   tt.fields.MsgCh,
				resultState:             tt.fields.resultState,
				lock:                    tt.fields.lock,
			}
			gotRsp, gotS := o.Preview(tt.args.request, tt.args.operation)
			if !reflect.DeepEqual(gotRsp, tt.wantRsp) {
				t.Errorf("Operation.Preview() gotRsp = %v, want %v", kdump.FormatN(gotRsp), kdump.FormatN(tt.wantRsp))
			}
			if !reflect.DeepEqual(gotS, tt.wantS) {
				t.Errorf("Operation.Preview() gotS = %v, want %v", gotS, tt.wantS)
			}
		})
	}
}

package operation

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"kusionstack.io/kusion/pkg/engine/states/local"

	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"

	"kusionstack.io/kusion/pkg/engine/operation/types"

	"kusionstack.io/kusion/pkg/engine/models"
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
		"models": map[string]interface{}{
			"type": "NodePort",
		},
	}
	FakeResourceState = models.Resource{
		ID:         "fake-id",
		Attributes: FakeService,
	}
	FakeResourceState2 = models.Resource{
		ID:         "fake-id-2",
		Attributes: FakeService,
	}
)

var _ runtime.Runtime = (*fakePreviewRuntime)(nil)

type fakePreviewRuntime struct{}

func (f *fakePreviewRuntime) Apply(ctx context.Context, request *runtime.ApplyRequest) *runtime.ApplyResponse {
	return &runtime.ApplyResponse{
		Resource: request.PlanResource,
		Status:   nil,
	}
}

func (f *fakePreviewRuntime) Read(ctx context.Context, request *runtime.ReadRequest) *runtime.ReadResponse {
	if request.Resource.ResourceKey() == "fake-id" {
		return &runtime.ReadResponse{
			Resource: nil,
			Status:   nil,
		}
	}
	return &runtime.ReadResponse{
		Resource: request.Resource,
		Status:   nil,
	}
}

func (f *fakePreviewRuntime) Delete(ctx context.Context, request *runtime.DeleteRequest) *runtime.DeleteResponse {
	return nil
}

func (f *fakePreviewRuntime) Watch(ctx context.Context, request *runtime.WatchRequest) *runtime.WatchResponse {
	return nil
}

func TestOperation_Preview(t *testing.T) {
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
		request *PreviewRequest
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
				OperationType: types.ApplyPreview,
				Runtime:       &fakePreviewRuntime{},
				StateStorage:  &local.FileSystemState{Path: local.KusionState},
				Order:         &opsmodels.ChangeOrder{StepKeys: []string{}, ChangeSteps: map[string]*opsmodels.ChangeStep{}},
			},
			args: args{
				request: &PreviewRequest{
					Request: opsmodels.Request{
						Tenant:   "fake-tenant",
						Stack:    "fake-stack",
						Project:  "fake-project",
						Operator: "fake-operator",
						Spec: &models.Spec{
							Resources: []models.Resource{
								FakeResourceState,
							},
						},
					},
				},
			},
			wantRsp: &PreviewResponse{
				Order: &opsmodels.ChangeOrder{
					StepKeys: []string{"fake-id"},
					ChangeSteps: map[string]*opsmodels.ChangeStep{
						"fake-id": {
							ID:       "fake-id",
							Action:   types.Create,
							Original: (*models.Resource)(nil),
							Modified: &FakeResourceState,
							Current:  (*models.Resource)(nil),
						},
					},
				},
			},
			wantS: nil,
		},
		{
			name: "success-when-destroy",
			fields: fields{
				OperationType: types.DestroyPreview,
				Runtime:       &fakePreviewRuntime{},
				StateStorage:  &local.FileSystemState{Path: local.KusionState},
				Order:         &opsmodels.ChangeOrder{},
			},
			args: args{
				request: &PreviewRequest{
					Request: opsmodels.Request{
						Tenant:   "fake-tenant",
						Stack:    "fake-stack",
						Project:  "fake-project",
						Operator: "fake-operator",
						Spec: &models.Spec{
							Resources: []models.Resource{
								FakeResourceState2,
							},
						},
					},
				},
			},
			wantRsp: &PreviewResponse{
				Order: &opsmodels.ChangeOrder{
					StepKeys: []string{"fake-id-2"},
					ChangeSteps: map[string]*opsmodels.ChangeStep{
						"fake-id-2": {
							ID:       "fake-id-2",
							Action:   types.Delete,
							Original: &FakeResourceState2,
							Modified: &FakeResourceState2,
							Current:  &FakeResourceState2,
						},
					},
				},
			},
			wantS: nil,
		},
		{
			name: "fail-because-empty-models",
			fields: fields{
				OperationType: types.ApplyPreview,
				Runtime:       &fakePreviewRuntime{},
				StateStorage:  &local.FileSystemState{Path: local.KusionState},
				Order:         &opsmodels.ChangeOrder{},
			},
			args: args{
				request: &PreviewRequest{
					Request: opsmodels.Request{
						Spec: nil,
					},
				},
			},
			wantRsp: nil,
			wantS:   status.NewErrorStatusWithMsg(status.InvalidArgument, "request.Spec is empty. If you want to delete all resources, please use command 'destroy'"),
		},
		{
			name: "fail-because-nonexistent-id",
			fields: fields{
				OperationType: types.ApplyPreview,
				Runtime:       &fakePreviewRuntime{},
				StateStorage:  &local.FileSystemState{Path: local.KusionState},
				Order:         &opsmodels.ChangeOrder{},
			},
			args: args{
				request: &PreviewRequest{
					Request: opsmodels.Request{
						Tenant:   "fake-tennat",
						Stack:    "fake-stack",
						Project:  "fake-project",
						Operator: "fake-operator",
						Spec: &models.Spec{
							Resources: []models.Resource{
								{
									ID: "fake-id",

									Attributes: FakeService,
									DependsOn:  []string{"nonexistent-id"},
								},
							},
						},
					},
				},
			},
			wantRsp: nil,
			wantS:   status.NewErrorStatusWithMsg(status.IllegalManifest, "can't find resource by key:nonexistent-id in models or state."),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &PreviewOperation{
				Operation: opsmodels.Operation{
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
				},
			}
			gotRsp, gotS := o.Preview(tt.args.request)
			if !reflect.DeepEqual(gotRsp, tt.wantRsp) {
				t.Errorf("Operation.Preview() gotRsp = %v, want %v", kdump.FormatN(gotRsp), kdump.FormatN(tt.wantRsp))
			}
			if !reflect.DeepEqual(gotS, tt.wantS) {
				t.Errorf("Operation.Preview() gotS = %v, want %v", gotS, tt.wantS)
			}
		})
	}
}

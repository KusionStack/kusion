package operation

import (
	"context"
	"reflect"
	"sync"
	"testing"

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
)

var _ runtime.Runtime = (*fakePreviewRuntime)(nil)

type fakePreviewRuntime struct{}

func (f *fakePreviewRuntime) Apply(ctx context.Context, priorState, planState *models.Resource) (*models.Resource, status.Status) {
	return planState, nil
}

func (f *fakePreviewRuntime) Read(ctx context.Context, resourceState *models.Resource) (*models.Resource, status.Status) {
	return resourceState, nil
}

func (f *fakePreviewRuntime) Delete(ctx context.Context, resourceState *models.Resource) status.Status {
	return nil
}

func (f *fakePreviewRuntime) Watch(ctx context.Context, resourceState *models.Resource) (*models.Resource, status.Status) {
	return resourceState, nil
}

func TestOperation_Preview(t *testing.T) {
	type fields struct {
		OperationType           Type
		StateStorage            states.StateStorage
		CtxResourceIndex        map[string]*models.Resource
		PriorStateResourceIndex map[string]*models.Resource
		StateResourceIndex      map[string]*models.Resource
		Order                   *ChangeOrder
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
				Runtime:      &fakePreviewRuntime{},
				StateStorage: &states.FileSystemState{Path: states.KusionState},
				Order:        &ChangeOrder{StepKeys: []string{}, ChangeSteps: map[string]*ChangeStep{}},
			},
			args: args{
				request: &PreviewRequest{
					Request: Request{
						Tenant:   "fake-tenant",
						Stack:    "fake-stack",
						Project:  "fake-project",
						Operator: "fake-operator",
						Manifest: &models.Spec{
							Resources: []models.Resource{
								FakeResourceState,
							},
						},
					},
				},
				operation: Apply,
			},
			wantRsp: &PreviewResponse{
				Order: &ChangeOrder{
					StepKeys: []string{"fake-id"},
					ChangeSteps: map[string]*ChangeStep{
						"fake-id": {
							ID:       "fake-id",
							Action:   Create,
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
				Runtime:      &fakePreviewRuntime{},
				StateStorage: &states.FileSystemState{Path: states.KusionState},
				Order:        &ChangeOrder{},
			},
			args: args{
				request: &PreviewRequest{
					Request: Request{
						Tenant:   "fake-tenant",
						Stack:    "fake-stack",
						Project:  "fake-project",
						Operator: "fake-operator",
						Manifest: &models.Spec{
							Resources: []models.Resource{
								FakeResourceState,
							},
						},
					},
				},
				operation: Destroy,
			},
			wantRsp: &PreviewResponse{
				Order: &ChangeOrder{
					StepKeys: []string{"fake-id"},
					ChangeSteps: map[string]*ChangeStep{
						"fake-id": {
							ID:       "fake-id",
							Action:   Delete,
							Original: &FakeResourceState,
							Modified: (*models.Resource)(nil),
							Current:  &FakeResourceState,
						},
					},
				},
			},
			wantS: nil,
		},
		{
			name: "fail-because-empty-models",
			fields: fields{
				Runtime:      &fakePreviewRuntime{},
				StateStorage: &states.FileSystemState{Path: states.KusionState},
				Order:        &ChangeOrder{},
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
			wantS:   status.NewErrorStatusWithMsg(status.InvalidArgument, "request.Spec is empty. If you want to delete all resources, please use command 'destroy'"),
		},
		{
			name: "fail-because-nonexistent-id",
			fields: fields{
				Runtime:      &fakePreviewRuntime{},
				StateStorage: &states.FileSystemState{Path: states.KusionState},
				Order:        &ChangeOrder{},
			},
			args: args{
				request: &PreviewRequest{
					Request: Request{
						Tenant:   "fake-tennat",
						Stack:    "fake-stack",
						Project:  "fake-project",
						Operator: "fake-operator",
						Manifest: &models.Spec{
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
				operation: Apply,
			},
			wantRsp: nil,
			wantS:   status.NewErrorStatusWithMsg(status.IllegalManifest, "can't find resource by key:nonexistent-id in models or state."),
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
				Order:                   tt.fields.Order,
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

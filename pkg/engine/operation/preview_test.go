package operation

import (
	"context"
	"os"
	"reflect"
	"sync"
	"testing"

	"bou.ke/monkey"

	"kusionstack.io/kusion/pkg/engine/models"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/engine/states/local"
	"kusionstack.io/kusion/pkg/projectstack"
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
		Type:       runtime.Kubernetes,
		Attributes: FakeService,
	}
	FakeResourceState2 = models.Resource{
		ID:         "fake-id-2",
		Type:       runtime.Kubernetes,
		Attributes: FakeService,
	}
)

var _ runtime.Runtime = (*fakePreviewRuntime)(nil)

type fakePreviewRuntime struct{}

func (f *fakePreviewRuntime) Import(ctx context.Context, request *runtime.ImportRequest) *runtime.ImportResponse {
	return &runtime.ImportResponse{Resource: request.PlanResource}
}

func (f *fakePreviewRuntime) Apply(ctx context.Context, request *runtime.ApplyRequest) *runtime.ApplyResponse {
	return &runtime.ApplyResponse{
		Resource: request.PlanResource,
		Status:   nil,
	}
}

func (f *fakePreviewRuntime) Read(ctx context.Context, request *runtime.ReadRequest) *runtime.ReadResponse {
	requestResource := request.PlanResource
	if requestResource == nil {
		requestResource = request.PriorResource
	}
	if requestResource.ResourceKey() == "fake-id" {
		return &runtime.ReadResponse{
			Resource: nil,
			Status:   nil,
		}
	}
	return &runtime.ReadResponse{
		Resource: requestResource,
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
	defer os.Remove("kusion_state.json")
	type fields struct {
		OperationType           opsmodels.OperationType
		StateStorage            states.StateStorage
		CtxResourceIndex        map[string]*models.Resource
		PriorStateResourceIndex map[string]*models.Resource
		StateResourceIndex      map[string]*models.Resource
		Order                   *opsmodels.ChangeOrder
		RuntimeMap              map[models.Type]runtime.Runtime
		MsgCh                   chan opsmodels.Message
		resultState             *states.State
		lock                    *sync.Mutex
	}
	type args struct {
		request *PreviewRequest
	}
	stack := &projectstack.Stack{
		StackConfiguration: projectstack.StackConfiguration{Name: "fake-name"},
		Path:               "fake-path",
	}
	project := &projectstack.Project{
		ProjectConfiguration: projectstack.ProjectConfiguration{
			Name:    "fake-name",
			Tenant:  "fake-tenant",
			Backend: nil,
		},
		Path:   "fake-path",
		Stacks: []*projectstack.Stack{stack},
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
				OperationType: opsmodels.ApplyPreview,
				RuntimeMap:    map[models.Type]runtime.Runtime{runtime.Kubernetes: &fakePreviewRuntime{}},
				StateStorage:  &local.FileSystemState{Path: local.KusionState},
				Order:         &opsmodels.ChangeOrder{StepKeys: []string{}, ChangeSteps: map[string]*opsmodels.ChangeStep{}},
			},
			args: args{
				request: &PreviewRequest{
					Request: opsmodels.Request{
						Tenant:   "fake-tenant",
						Stack:    stack,
						Project:  project,
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
							ID:     "fake-id",
							Action: opsmodels.Create,
							From:   (*models.Resource)(nil),
							To:     &FakeResourceState,
						},
					},
				},
			},
			wantS: nil,
		},
		{
			name: "success-when-destroy",
			fields: fields{
				OperationType: opsmodels.DestroyPreview,
				RuntimeMap:    map[models.Type]runtime.Runtime{runtime.Kubernetes: &fakePreviewRuntime{}},
				StateStorage:  &local.FileSystemState{Path: local.KusionState},
				Order:         &opsmodels.ChangeOrder{},
			},
			args: args{
				request: &PreviewRequest{
					Request: opsmodels.Request{
						Tenant:   "fake-tenant",
						Stack:    stack,
						Project:  project,
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
							ID:     "fake-id-2",
							Action: opsmodels.Delete,
							From:   &FakeResourceState2,
							To:     &FakeResourceState2,
						},
					},
				},
			},
			wantS: nil,
		},
		{
			name: "fail-because-empty-models",
			fields: fields{
				OperationType: opsmodels.ApplyPreview,
				RuntimeMap:    map[models.Type]runtime.Runtime{runtime.Kubernetes: &fakePreviewRuntime{}},
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
				OperationType: opsmodels.ApplyPreview,
				RuntimeMap:    map[models.Type]runtime.Runtime{runtime.Kubernetes: &fakePreviewRuntime{}},
				StateStorage:  &local.FileSystemState{Path: local.KusionState},
				Order:         &opsmodels.ChangeOrder{},
			},
			args: args{
				request: &PreviewRequest{
					Request: opsmodels.Request{
						Tenant:   "fake-tenant",
						Stack:    stack,
						Project:  project,
						Operator: "fake-operator",
						Spec: &models.Spec{
							Resources: []models.Resource{
								{
									ID:         "fake-id",
									Type:       runtime.Kubernetes,
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
					RuntimeMap:              tt.fields.RuntimeMap,
					MsgCh:                   tt.fields.MsgCh,
					ResultState:             tt.fields.resultState,
					Lock:                    tt.fields.lock,
				},
			}

			monkey.Patch(runtimeinit.Runtimes, func(resources models.Resources) (map[models.Type]runtime.Runtime, status.Status) {
				return map[models.Type]runtime.Runtime{runtime.Kubernetes: &fakePreviewRuntime{}}, nil
			})
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

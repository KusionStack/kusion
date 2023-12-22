package operation

import (
	"context"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/bytedance/mockey"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/engine/states/local"
	jsonutil "kusionstack.io/kusion/pkg/util/json"
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
	FakeResourceState = apiv1.Resource{
		ID:         "fake-id",
		Type:       runtime.Kubernetes,
		Attributes: FakeService,
	}
	FakeResourceState2 = apiv1.Resource{
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
		CtxResourceIndex        map[string]*apiv1.Resource
		PriorStateResourceIndex map[string]*apiv1.Resource
		StateResourceIndex      map[string]*apiv1.Resource
		Order                   *opsmodels.ChangeOrder
		RuntimeMap              map[apiv1.Type]runtime.Runtime
		MsgCh                   chan opsmodels.Message
		resultState             *states.State
		lock                    *sync.Mutex
	}
	type args struct {
		request *PreviewRequest
	}
	s := &apiv1.Stack{
		Name: "fake-name",
		Path: "fake-path",
	}
	p := &apiv1.Project{
		Name:   "fake-name",
		Path:   "fake-path",
		Stacks: []*apiv1.Stack{s},
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRsp *PreviewResponse
		wantS   v1.Status
	}{
		{
			name: "success-when-apply",
			fields: fields{
				OperationType: opsmodels.ApplyPreview,
				RuntimeMap:    map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: &fakePreviewRuntime{}},
				StateStorage:  &local.FileSystemState{Path: local.KusionStateFileFile},
				Order:         &opsmodels.ChangeOrder{StepKeys: []string{}, ChangeSteps: map[string]*opsmodels.ChangeStep{}},
			},
			args: args{
				request: &PreviewRequest{
					Request: opsmodels.Request{
						Tenant:   "fake-tenant",
						Stack:    s,
						Project:  p,
						Operator: "fake-operator",
						Intent: &apiv1.Intent{
							Resources: []apiv1.Resource{
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
							From:   (*apiv1.Resource)(nil),
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
				RuntimeMap:    map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: &fakePreviewRuntime{}},
				StateStorage:  &local.FileSystemState{Path: local.KusionStateFileFile},
				Order:         &opsmodels.ChangeOrder{},
			},
			args: args{
				request: &PreviewRequest{
					Request: opsmodels.Request{
						Tenant:   "fake-tenant",
						Stack:    s,
						Project:  p,
						Operator: "fake-operator",
						Intent: &apiv1.Intent{
							Resources: []apiv1.Resource{
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
							To:     (*apiv1.Resource)(nil),
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
				RuntimeMap:    map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: &fakePreviewRuntime{}},
				StateStorage:  &local.FileSystemState{Path: local.KusionStateFileFile},
				Order:         &opsmodels.ChangeOrder{},
			},
			args: args{
				request: &PreviewRequest{
					Request: opsmodels.Request{
						Intent: nil,
					},
				},
			},
			wantRsp: nil,
			wantS:   v1.NewErrorStatusWithMsg(v1.InvalidArgument, "request.Intent is empty. If you want to delete all resources, please use command 'destroy'"),
		},
		{
			name: "fail-because-nonexistent-id",
			fields: fields{
				OperationType: opsmodels.ApplyPreview,
				RuntimeMap:    map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: &fakePreviewRuntime{}},
				StateStorage:  &local.FileSystemState{Path: local.KusionStateFileFile},
				Order:         &opsmodels.ChangeOrder{},
			},
			args: args{
				request: &PreviewRequest{
					Request: opsmodels.Request{
						Tenant:   "fake-tenant",
						Stack:    s,
						Project:  p,
						Operator: "fake-operator",
						Intent: &apiv1.Intent{
							Resources: []apiv1.Resource{
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
			wantS:   v1.NewErrorStatusWithMsg(v1.IllegalManifest, "can't find resource by key:nonexistent-id in models or state."),
		},
	}
	for _, tt := range tests {
		mockey.PatchConvey(tt.name, t, func() {
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

			mockey.Mock(runtimeinit.Runtimes).To(func(
				resources apiv1.Resources,
			) (map[apiv1.Type]runtime.Runtime, v1.Status) {
				return map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: &fakePreviewRuntime{}}, nil
			}).Build()
			gotRsp, gotS := o.Preview(tt.args.request)
			if !reflect.DeepEqual(gotRsp, tt.wantRsp) {
				t.Errorf("Operation.Preview() gotRsp = %v, want %v", jsonutil.Marshal2PrettyString(gotRsp), jsonutil.Marshal2PrettyString(tt.wantRsp))
			}
			if !reflect.DeepEqual(gotS, tt.wantS) {
				t.Errorf("Operation.Preview() gotS = %v, want %v", gotS, tt.wantS)
			}
		})
	}
}

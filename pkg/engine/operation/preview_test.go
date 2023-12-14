package operation

import (
	"context"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/bytedance/mockey"

	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/project"
	"kusionstack.io/kusion/pkg/apis/stack"
	"kusionstack.io/kusion/pkg/apis/status"
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
	FakeResourceState = intent.Resource{
		ID:         "fake-id",
		Type:       runtime.Kubernetes,
		Attributes: FakeService,
	}
	FakeResourceState2 = intent.Resource{
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
		CtxResourceIndex        map[string]*intent.Resource
		PriorStateResourceIndex map[string]*intent.Resource
		StateResourceIndex      map[string]*intent.Resource
		Order                   *opsmodels.ChangeOrder
		RuntimeMap              map[intent.Type]runtime.Runtime
		MsgCh                   chan opsmodels.Message
		resultState             *states.State
		lock                    *sync.Mutex
	}
	type args struct {
		request *PreviewRequest
	}
	s := &stack.Stack{
		Configuration: stack.Configuration{Name: "fake-name"},
		Path:          "fake-path",
	}
	p := &project.Project{
		Configuration: project.Configuration{
			Name:   "fake-name",
			Tenant: "fake-tenant",
		},
		Path:   "fake-path",
		Stacks: []*stack.Stack{s},
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
				RuntimeMap:    map[intent.Type]runtime.Runtime{runtime.Kubernetes: &fakePreviewRuntime{}},
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
						Intent: &intent.Intent{
							Resources: []intent.Resource{
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
							From:   (*intent.Resource)(nil),
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
				RuntimeMap:    map[intent.Type]runtime.Runtime{runtime.Kubernetes: &fakePreviewRuntime{}},
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
						Intent: &intent.Intent{
							Resources: []intent.Resource{
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
							To:     (*intent.Resource)(nil),
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
				RuntimeMap:    map[intent.Type]runtime.Runtime{runtime.Kubernetes: &fakePreviewRuntime{}},
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
			wantS:   status.NewErrorStatusWithMsg(status.InvalidArgument, "request.Intent is empty. If you want to delete all resources, please use command 'destroy'"),
		},
		{
			name: "fail-because-nonexistent-id",
			fields: fields{
				OperationType: opsmodels.ApplyPreview,
				RuntimeMap:    map[intent.Type]runtime.Runtime{runtime.Kubernetes: &fakePreviewRuntime{}},
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
						Intent: &intent.Intent{
							Resources: []intent.Resource{
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
				resources intent.Resources,
			) (map[intent.Type]runtime.Runtime, status.Status) {
				return map[intent.Type]runtime.Runtime{runtime.Kubernetes: &fakePreviewRuntime{}}, nil
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

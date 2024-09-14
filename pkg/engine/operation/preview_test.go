package operation

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/bytedance/mockey"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/engine/release/storages"
	"kusionstack.io/kusion/pkg/engine/runtime"
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/pkg/util/json"
)

var _ runtime.Runtime = (*fakePreviewRuntime)(nil)

type fakePreviewRuntime struct{}

func (f *fakePreviewRuntime) Import(_ context.Context, request *runtime.ImportRequest) *runtime.ImportResponse {
	return &runtime.ImportResponse{Resource: request.PlanResource}
}

func (f *fakePreviewRuntime) Apply(_ context.Context, request *runtime.ApplyRequest) *runtime.ApplyResponse {
	return &runtime.ApplyResponse{
		Resource: request.PlanResource,
		Status:   nil,
	}
}

func (f *fakePreviewRuntime) Read(_ context.Context, request *runtime.ReadRequest) *runtime.ReadResponse {
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

func (f *fakePreviewRuntime) Delete(_ context.Context, _ *runtime.DeleteRequest) *runtime.DeleteResponse {
	return nil
}

func (f *fakePreviewRuntime) Watch(_ context.Context, _ *runtime.WatchRequest) *runtime.WatchResponse {
	return nil
}

func TestPreviewOperation_Preview(t *testing.T) {
	type fields struct {
		operationType           models.OperationType
		releaseStorage          release.Storage
		ctxResourceIndex        map[string]*apiv1.Resource
		priorStateResourceIndex map[string]*apiv1.Resource
		stateResourceIndex      map[string]*apiv1.Resource
		order                   *models.ChangeOrder
		runtimeMap              map[apiv1.Type]runtime.Runtime
		msgCh                   chan models.Message
		release                 *apiv1.Release
		lock                    *sync.Mutex
	}
	type args struct {
		req *PreviewRequest
	}

	fakeStack := &apiv1.Stack{
		Name: "fake-stack",
		Path: "fake-path",
	}
	fakeProject := &apiv1.Project{
		Name:   "fake-project",
		Path:   "fake-path",
		Stacks: []*apiv1.Stack{fakeStack},
	}
	fakeService := map[string]interface{}{
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
	fakeResource := apiv1.Resource{
		ID:         "fake-id",
		Type:       runtime.Kubernetes,
		Attributes: fakeService,
	}
	fakeResource2 := apiv1.Resource{
		ID:         "fake-id-2",
		Type:       runtime.Kubernetes,
		Attributes: fakeService,
	}

	testcases := []struct {
		name    string
		fields  fields
		args    args
		wantRsp *PreviewResponse
		wantErr bool
	}{
		{
			name: "success-when-apply",
			fields: fields{
				operationType:  models.ApplyPreview,
				runtimeMap:     map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: &fakePreviewRuntime{}},
				releaseStorage: &storages.LocalStorage{},
				order: &models.ChangeOrder{
					StepKeys:    []string{},
					ChangeSteps: map[string]*models.ChangeStep{},
				},
			},
			args: args{
				req: &PreviewRequest{
					Request: models.Request{
						Stack:   fakeStack,
						Project: fakeProject,
					},
					Spec: &apiv1.Spec{
						Resources: apiv1.Resources{fakeResource},
					},
					State: &apiv1.State{},
				},
			},
			wantRsp: &PreviewResponse{
				Order: &models.ChangeOrder{
					StepKeys: []string{"fake-id"},
					ChangeSteps: map[string]*models.ChangeStep{
						"fake-id": {
							ID:     "fake-id",
							Action: models.Create,
							From:   (*apiv1.Resource)(nil),
							To:     &fakeResource,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success-when-destroy",
			fields: fields{
				operationType:  models.DestroyPreview,
				runtimeMap:     map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: &fakePreviewRuntime{}},
				releaseStorage: &storages.LocalStorage{},
				order:          &models.ChangeOrder{},
			},
			args: args{
				req: &PreviewRequest{
					Request: models.Request{
						Stack:   fakeStack,
						Project: fakeProject,
					},
					Spec: &apiv1.Spec{
						Resources: apiv1.Resources{fakeResource2},
					},
					State: &apiv1.State{},
				},
			},
			wantRsp: &PreviewResponse{
				Order: &models.ChangeOrder{
					StepKeys: []string{"fake-id-2"},
					ChangeSteps: map[string]*models.ChangeStep{
						"fake-id-2": {
							ID:     "fake-id-2",
							Action: models.Delete,
							From:   &fakeResource2,
							To:     (*apiv1.Resource)(nil),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "fail-because-empty-models",
			fields: fields{
				operationType:  models.ApplyPreview,
				runtimeMap:     map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: &fakePreviewRuntime{}},
				releaseStorage: &storages.LocalStorage{},
				order:          &models.ChangeOrder{},
			},
			args: args{
				req: &PreviewRequest{
					Spec:  nil,
					State: &apiv1.State{},
				},
			},
			wantRsp: nil,
			wantErr: true,
		},
		{
			name: "fail-because-nonexistent-id",
			fields: fields{
				operationType:  models.ApplyPreview,
				runtimeMap:     map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: &fakePreviewRuntime{}},
				releaseStorage: &storages.LocalStorage{},
				order:          &models.ChangeOrder{},
			},
			args: args{
				req: &PreviewRequest{
					Request: models.Request{
						Stack:   fakeStack,
						Project: fakeProject,
					},
					Spec: &apiv1.Spec{
						Resources: []apiv1.Resource{
							{
								ID:         "fake-id",
								Type:       runtime.Kubernetes,
								Attributes: fakeService,
								DependsOn:  []string{"nonexistent-id"},
							},
						},
					},
					State: &apiv1.State{},
				},
			},
			wantRsp: nil,
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		mockey.PatchConvey(tc.name, t, func() {
			o := &PreviewOperation{
				Operation: models.Operation{
					OperationType:           tc.fields.operationType,
					ReleaseStorage:          tc.fields.releaseStorage,
					CtxResourceIndex:        tc.fields.ctxResourceIndex,
					PriorStateResourceIndex: tc.fields.priorStateResourceIndex,
					StateResourceIndex:      tc.fields.stateResourceIndex,
					ChangeOrder:             tc.fields.order,
					RuntimeMap:              tc.fields.runtimeMap,
					MsgCh:                   tc.fields.msgCh,
					Release:                 tc.fields.release,
					Lock:                    tc.fields.lock,
				},
			}

			mockey.Mock(runtimeinit.Runtimes).To(func(
				spec apiv1.Spec, state apiv1.State,
			) (map[apiv1.Type]runtime.Runtime, v1.Status) {
				return map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: &fakePreviewRuntime{}}, nil
			}).Build()

			gotRsp, gotS := o.Preview(tc.args.req)
			if !reflect.DeepEqual(gotRsp, tc.wantRsp) {
				t.Errorf("Operation.Preview() gotRsp = %v, want %v", json.Marshal2PrettyString(gotRsp), json.Marshal2PrettyString(tc.wantRsp))
			}
			if tc.wantErr && gotS == nil || !tc.wantErr && gotS != nil {
				t.Errorf("Operation.Preview() gotS = %v, wantErr %v", gotS, tc.wantErr)
			}
		})
	}
}

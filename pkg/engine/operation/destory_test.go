package operation

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation/graph"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/engine/release/storages"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
)

var _ runtime.Runtime = (*fakerRuntime)(nil)

type fakerRuntime struct{}

func (f *fakerRuntime) Import(_ context.Context, request *runtime.ImportRequest) *runtime.ImportResponse {
	return &runtime.ImportResponse{Resource: request.PlanResource}
}

func (f *fakerRuntime) Apply(_ context.Context, request *runtime.ApplyRequest) *runtime.ApplyResponse {
	return &runtime.ApplyResponse{
		Resource: request.PlanResource,
		Status:   nil,
	}
}

func (f *fakerRuntime) Read(_ context.Context, request *runtime.ReadRequest) *runtime.ReadResponse {
	if request.PlanResource.ResourceKey() == "fake-id" {
		return &runtime.ReadResponse{
			Resource: nil,
			Status:   nil,
		}
	}
	return &runtime.ReadResponse{
		Resource: request.PlanResource,
		Status:   nil,
	}
}

func (f *fakerRuntime) Delete(_ context.Context, _ *runtime.DeleteRequest) *runtime.DeleteResponse {
	return nil
}

func (f *fakerRuntime) Watch(_ context.Context, _ *runtime.WatchRequest) *runtime.WatchResponse {
	return nil
}

func TestDestroyOperation_Destroy(t *testing.T) {
	fakeStack := &apiv1.Stack{
		Name: "fake-project",
		Path: "fake-path",
	}
	fakeProject := &apiv1.Project{
		Name:   "fake-stack",
		Path:   "fake-path",
		Stacks: []*apiv1.Stack{fakeStack},
	}

	fakeResource := apiv1.Resource{
		ID:   "id1",
		Type: runtime.Kubernetes,
		Attributes: map[string]interface{}{
			"foo": "bar",
		},
		DependsOn: nil,
	}
	fakeSpec := &apiv1.Spec{Resources: []apiv1.Resource{fakeResource}}
	fakeState := &apiv1.State{Resources: []apiv1.Resource{fakeResource}}

	loc, _ := time.LoadLocation("Asia/Shanghai")
	fakeTime := time.Date(2024, 5, 20, 14, 51, 0, 0, loc)
	fakeRelease := &apiv1.Release{
		Project:      "fake-project",
		Workspace:    "fake-workspace",
		Revision:     2,
		Stack:        "fake-stack",
		Spec:         fakeSpec,
		State:        fakeState,
		Phase:        apiv1.ReleasePhaseDestroying,
		CreateTime:   fakeTime,
		ModifiedTime: fakeTime,
	}
	fakeDestroyRelease := &apiv1.Release{
		Project:   "fake-project",
		Workspace: "fake-workspace",
		Revision:  2,
		Stack:     "fake-stack",
		Spec:      nil,
		State: &apiv1.State{
			Resources: apiv1.Resources{},
		},
		Phase:        apiv1.ReleasePhaseDestroying,
		CreateTime:   fakeTime,
		ModifiedTime: fakeTime,
	}

	o := &DestroyOperation{
		models.Operation{
			OperationType:  models.Destroy,
			ReleaseStorage: &storages.LocalStorage{},
			RuntimeMap:     map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: &kubernetes.KubernetesRuntime{}},
		},
	}
	req := &DestroyRequest{
		Request: models.Request{
			Stack:   fakeStack,
			Project: fakeProject,
		},
		Release: fakeRelease,
	}
	expectedRsp := &DestroyResponse{
		Release: fakeDestroyRelease,
	}

	mockey.PatchConvey("destroy success", t, func() {
		mockey.Mock((*graph.ResourceNode).Execute).To(func(operation *models.Operation) v1.Status {
			operation.Release = fakeDestroyRelease
			return nil
		}).Build()
		mockey.Mock(kubernetes.NewKubernetesRuntime).To(func() (runtime.Runtime, error) {
			return &fakerRuntime{}, nil
		}).Build()

		o.MsgCh = make(chan models.Message, 1)
		go readMsgCh(o.MsgCh)
		rsp, status := o.Destroy(req)
		assert.Equal(t, rsp, expectedRsp)
		assert.Nil(t, status)
	})

	mockey.PatchConvey("destroy failed", t, func() {
		mockey.Mock((*graph.ResourceNode).Execute).Return(v1.NewErrorStatus(errors.New("mock error"))).Build()
		mockey.Mock(kubernetes.NewKubernetesRuntime).Return(&fakerRuntime{}, nil).Build()

		o.MsgCh = make(chan models.Message, 1)
		go readMsgCh(o.MsgCh)
		_, status := o.Destroy(req)
		assert.True(t, v1.IsErr(status))
	})
}

func readMsgCh(ch chan models.Message) {
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Printf("msg: %+v\n", msg)
		}
	}
}

func Test_ValidateDestroyRequest(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		req     *DestroyRequest
	}{
		{
			name:    "invalid request nil request",
			success: false,
			req:     nil,
		},
		{
			name:    "invalid request invalid release phase",
			success: false,
			req: &DestroyRequest{
				Release: &apiv1.Release{
					Phase: "invalid_phase",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock valid release and spec", t, func() {
				mockey.Mock(release.ValidateRelease).Return(nil).Build()
				mockey.Mock(release.ValidateSpec).Return(nil).Build()
				err := validateDestroyRequest(tc.req)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

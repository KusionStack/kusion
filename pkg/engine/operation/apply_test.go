package operation

import (
	"sync"
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
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
)

func TestApplyOperation_Apply(t *testing.T) {
	type fields struct {
		operationType           models.OperationType
		releaseStorage          release.Storage
		ctxResourceIndex        map[string]*apiv1.Resource
		priorStateResourceIndex map[string]*apiv1.Resource
		stateResourceIndex      map[string]*apiv1.Resource
		order                   *models.ChangeOrder
		runtimeMap              map[apiv1.Type]runtime.Runtime
		stack                   *apiv1.Stack
		msgCh                   chan models.Message
		release                 *apiv1.Release
		lock                    *sync.Mutex
	}
	type args struct {
		applyRequest *ApplyRequest
	}

	fakeSpec := &apiv1.Spec{
		Resources: []apiv1.Resource{
			{
				ID:   "mock-id",
				Type: runtime.Kubernetes,
				Attributes: map[string]interface{}{
					"a": "b",
				},
				DependsOn: nil,
			},
		},
	}
	fakeState := &apiv1.State{
		Resources: []apiv1.Resource{
			{
				ID:   "mock-id",
				Type: runtime.Kubernetes,
				Attributes: map[string]interface{}{
					"a": "b",
				},
				DependsOn: nil,
			},
		},
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")
	fakeTime := time.Date(2024, 5, 10, 16, 48, 0, 0, loc)
	fakeRelease := &apiv1.Release{
		Project:      "fake-project",
		Workspace:    "fake-workspace",
		Revision:     1,
		Stack:        "fake-stack",
		Spec:         fakeSpec,
		State:        &apiv1.State{},
		Phase:        apiv1.ReleasePhaseApplying,
		CreateTime:   fakeTime,
		ModifiedTime: fakeTime,
	}
	fakeUpdateRelease := &apiv1.Release{
		Project:      "fake-project",
		Workspace:    "fake-workspace",
		Revision:     1,
		Stack:        "fake-stack",
		Spec:         fakeSpec,
		State:        fakeState,
		Phase:        apiv1.ReleasePhaseApplying,
		CreateTime:   fakeTime,
		ModifiedTime: fakeTime,
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

	testcases := []struct {
		name             string
		fields           fields
		args             args
		expectedResponse *ApplyResponse
		expectedStatus   v1.Status
	}{
		{
			name: "apply test",
			fields: fields{
				operationType:  models.Apply,
				releaseStorage: &storages.LocalStorage{},
				runtimeMap:     map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: &kubernetes.KubernetesRuntime{}},
				msgCh:          make(chan models.Message, 5),
			},
			args: args{applyRequest: &ApplyRequest{
				Request: models.Request{
					Stack:   fakeStack,
					Project: fakeProject,
				},
				Release: fakeRelease,
			}},
			expectedResponse: &ApplyResponse{Release: fakeUpdateRelease},
			expectedStatus:   nil,
		},
	}

	for _, tc := range testcases {
		mockey.PatchConvey(tc.name, t, func() {
			o := &models.Operation{
				OperationType:           tc.fields.operationType,
				ReleaseStorage:          tc.fields.releaseStorage,
				CtxResourceIndex:        tc.fields.ctxResourceIndex,
				PriorStateResourceIndex: tc.fields.priorStateResourceIndex,
				StateResourceIndex:      tc.fields.stateResourceIndex,
				ChangeOrder:             tc.fields.order,
				RuntimeMap:              tc.fields.runtimeMap,
				Stack:                   tc.fields.stack,
				MsgCh:                   tc.fields.msgCh,
				Release:                 tc.fields.release,
				Lock:                    tc.fields.lock,
			}
			ao := &ApplyOperation{
				Operation: *o,
			}

			mockey.Mock((*graph.ResourceNode).Execute).To(func(operation *models.Operation) v1.Status {
				operation.Release = fakeUpdateRelease
				return nil
			}).Build()
			mockey.Mock(runtimeinit.Runtimes).To(func(
				spec apiv1.Spec, state apiv1.State,
			) (map[apiv1.Type]runtime.Runtime, v1.Status) {
				return map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: &kubernetes.KubernetesRuntime{}}, nil
			}).Build()

			rsp, status := ao.Apply(tc.args.applyRequest)
			assert.Equal(t, tc.expectedResponse, rsp)
			assert.Equal(t, tc.expectedStatus, status)
		})
	}
}

func Test_ValidateApplyRequest(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		req     *ApplyRequest
	}{
		{
			name:    "invalid request nil request",
			success: false,
			req:     nil,
		},
		{
			name:    "invalid request invalid release phase",
			success: false,
			req: &ApplyRequest{
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
				err := validateApplyRequest(tc.req)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

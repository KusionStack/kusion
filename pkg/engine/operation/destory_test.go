//go:build !arm64
// +build !arm64

package operation

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/engine/operation/graph"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/engine/states/local"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/status"
)

func TestOperation_Destroy(t *testing.T) {
	var (
		tenant   = "tenant_name"
		operator = "foo_user"
	)

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

	resourceState := models.Resource{
		ID:   "id1",
		Type: runtime.Kubernetes,
		Attributes: map[string]interface{}{
			"foo": "bar",
		},
		DependsOn: nil,
	}
	mf := &models.Spec{Resources: []models.Resource{resourceState}}
	o := &DestroyOperation{
		opsmodels.Operation{
			OperationType: opsmodels.Destroy,
			StateStorage:  &local.FileSystemState{Path: filepath.Join("test_data", local.KusionState)},
			RuntimeMap:    map[models.Type]runtime.Runtime{runtime.Kubernetes: &kubernetes.KubernetesRuntime{}},
		},
	}
	r := &DestroyRequest{
		opsmodels.Request{
			Tenant:   tenant,
			Stack:    stack,
			Project:  project,
			Operator: operator,
			Spec:     mf,
		},
	}

	mockey.PatchConvey("destroy success", t, func() {
		mockey.Mock((*graph.ResourceNode).Execute).To(func(rn *graph.ResourceNode, operation *opsmodels.Operation) status.Status {
			return nil
		}).Build()
		mockey.Mock(mockey.GetMethod(local.NewFileSystemState(), "GetLatestState")).To(func(
			f *local.FileSystemState,
			query *states.StateQuery,
		) (*states.State, error) {
			return &states.State{Resources: []models.Resource{resourceState}}, nil
		}).Build()
		mockey.Mock(kubernetes.NewKubernetesRuntime).To(func() (runtime.Runtime, error) {
			return &fakerRuntime{}, nil
		}).Build()

		o.MsgCh = make(chan opsmodels.Message, 1)
		go readMsgCh(o.MsgCh)
		st := o.Destroy(r)
		assert.Nil(t, st)
	})

	mockey.PatchConvey("destroy failed", t, func() {
		mockey.Mock((*graph.ResourceNode).Execute).To(func(rn *graph.ResourceNode, operation *opsmodels.Operation) status.Status {
			return status.NewErrorStatus(errors.New("mock error"))
		}).Build()
		mockey.Mock(mockey.GetMethod(local.NewFileSystemState(), "GetLatestState")).To(func(
			f *local.FileSystemState,
			query *states.StateQuery,
		) (*states.State, error) {
			return &states.State{Resources: []models.Resource{resourceState}}, nil
		}).Build()
		mockey.Mock(kubernetes.NewKubernetesRuntime).To(func() (runtime.Runtime, error) {
			return &fakerRuntime{}, nil
		}).Build()

		o.MsgCh = make(chan opsmodels.Message, 1)
		go readMsgCh(o.MsgCh)
		st := o.Destroy(r)
		assert.True(t, status.IsErr(st))
	})
}

func readMsgCh(ch chan opsmodels.Message) {
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

var _ runtime.Runtime = (*fakerRuntime)(nil)

type fakerRuntime struct{}

func (f *fakerRuntime) Import(ctx context.Context, request *runtime.ImportRequest) *runtime.ImportResponse {
	return &runtime.ImportResponse{Resource: request.PlanResource}
}

func (f *fakerRuntime) Apply(ctx context.Context, request *runtime.ApplyRequest) *runtime.ApplyResponse {
	return &runtime.ApplyResponse{
		Resource: request.PlanResource,
		Status:   nil,
	}
}

func (f *fakerRuntime) Read(ctx context.Context, request *runtime.ReadRequest) *runtime.ReadResponse {
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

func (f *fakerRuntime) Delete(ctx context.Context, request *runtime.DeleteRequest) *runtime.DeleteResponse {
	return nil
}

func (f *fakerRuntime) Watch(ctx context.Context, request *runtime.WatchRequest) *runtime.WatchResponse {
	return nil
}

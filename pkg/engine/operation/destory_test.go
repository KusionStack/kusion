package operation

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation/graph"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
	"kusionstack.io/kusion/pkg/engine/state/storages"
)

func TestOperation_Destroy(t *testing.T) {
	operator := "foo-user"

	s := &apiv1.Stack{
		Name: "fake-name",
		Path: "fake-path",
	}
	p := &apiv1.Project{
		Name:   "fake-name",
		Path:   "fake-path",
		Stacks: []*apiv1.Stack{s},
	}

	resourceState := apiv1.Resource{
		ID:   "id1",
		Type: runtime.Kubernetes,
		Attributes: map[string]interface{}{
			"foo": "bar",
		},
		DependsOn: nil,
	}
	mf := &apiv1.Spec{Resources: []apiv1.Resource{resourceState}}
	o := &DestroyOperation{
		models.Operation{
			OperationType: models.Destroy,
			StateStorage:  storages.NewLocalStorage(filepath.Join("testdata", "state.yaml")),
			RuntimeMap:    map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: &kubernetes.KubernetesRuntime{}},
		},
	}
	r := &DestroyRequest{
		models.Request{
			Stack:    s,
			Project:  p,
			Operator: operator,
			Intent:   mf,
		},
	}

	mockey.PatchConvey("destroy success", t, func() {
		mockey.Mock((*graph.ResourceNode).Execute).Return(nil).Build()
		mockey.Mock((*storages.LocalStorage).Get).Return(&apiv1.DeprecatedState{Resources: []apiv1.Resource{resourceState}}, nil).Build()
		mockey.Mock(kubernetes.NewKubernetesRuntime).To(func() (runtime.Runtime, error) {
			return &fakerRuntime{}, nil
		}).Build()

		o.MsgCh = make(chan models.Message, 1)
		go readMsgCh(o.MsgCh)
		st := o.Destroy(r)
		assert.Nil(t, st)
	})

	mockey.PatchConvey("destroy failed", t, func() {
		mockey.Mock((*graph.ResourceNode).Execute).Return(v1.NewErrorStatus(errors.New("mock error"))).Build()
		mockey.Mock((*storages.LocalStorage).Get).Return(&apiv1.DeprecatedState{Resources: []apiv1.Resource{resourceState}}, nil).Build()
		mockey.Mock(kubernetes.NewKubernetesRuntime).Return(&fakerRuntime{}, nil).Build()

		o.MsgCh = make(chan models.Message, 1)
		go readMsgCh(o.MsgCh)
		st := o.Destroy(r)
		assert.True(t, v1.IsErr(st))
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

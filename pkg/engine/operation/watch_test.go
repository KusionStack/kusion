package operation

import (
	"context"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
)

var barDeployment = map[string]interface{}{
	"apiVersion": "apps/v1",
	"kind":       "Deployment",
	"metadata": map[string]interface{}{
		"namespace": "foo",
		"name":      "bar",
	},
	"spec": map[string]interface{}{
		"replica": 1,
		"template": map[string]interface{}{
			"spec": map[string]interface{}{
				"containers": []map[string]interface{}{
					{
						"image": "foo.bar.com:v1",
						"name":  "bar",
					},
				},
			},
		},
	},
}

var (
	fooRuntime                 = &fooWatchRuntime{}
	_          runtime.Runtime = (*fooWatchRuntime)(nil)
)

type fooWatchRuntime struct{}

func (f *fooWatchRuntime) Import(_ context.Context, request *runtime.ImportRequest) *runtime.ImportResponse {
	return &runtime.ImportResponse{Resource: request.PlanResource}
}

func (f *fooWatchRuntime) Apply(_ context.Context, _ *runtime.ApplyRequest) *runtime.ApplyResponse {
	return nil
}

func (f *fooWatchRuntime) Read(_ context.Context, _ *runtime.ReadRequest) *runtime.ReadResponse {
	return nil
}

func (f *fooWatchRuntime) Delete(_ context.Context, _ *runtime.DeleteRequest) *runtime.DeleteResponse {
	return nil
}

func (f *fooWatchRuntime) Watch(_ context.Context, _ *runtime.WatchRequest) *runtime.WatchResponse {
	out := make(chan watch.Event)
	go func() {
		out <- watch.Event{
			Type:   watch.Deleted,
			Object: &unstructured.Unstructured{Object: barDeployment},
		}
		close(out)
	}()

	return &runtime.WatchResponse{
		Watchers: &runtime.SequentialWatchers{
			IDs:      []string{"apps/v1:Deployment:foo:bar"},
			Watchers: []<-chan watch.Event{out},
		},
		Status: nil,
	}
}

func TestWatchOperation_Watch(t *testing.T) {
	mockey.PatchConvey("test watch operation: watch", t, func() {
		req := &WatchRequest{
			Spec: &apiv1.Spec{
				Resources: apiv1.Resources{
					{
						ID:         "apps/v1:Deployment:foo:bar",
						Type:       runtime.Kubernetes,
						Attributes: barDeployment,
					},
				},
			},
		}
		mockey.Mock(runtimeinit.Runtimes).To(func(
			resources apiv1.Resources,
		) (map[apiv1.Type]runtime.Runtime, v1.Status) {
			return map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: fooRuntime}, nil
		}).Build()
		wo := &WatchOperation{models.Operation{RuntimeMap: map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: fooRuntime}}}
		err := wo.Watch(req)
		assert.Nil(t, err)
	})
}

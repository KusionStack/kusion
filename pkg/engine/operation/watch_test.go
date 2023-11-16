package operation

import (
	"context"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sWatch "k8s.io/apimachinery/pkg/watch"

	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/status"
)

func TestWatchOperation_Watch(t *testing.T) {
	mockey.PatchConvey("test watch operation: watch", t, func() {
		req := &WatchRequest{
			Request: opsmodels.Request{
				Spec: &models.Intent{
					Resources: models.Resources{
						{
							ID:         "apps/v1:Deployment:foo:bar",
							Type:       runtime.Kubernetes,
							Attributes: barDeployment,
						},
					},
				},
			},
		}
		mockey.Mock(runtimeinit.Runtimes).To(func(
			resources models.Resources,
			stack *projectstack.Stack,
		) (map[models.Type]runtime.Runtime, status.Status) {
			return map[models.Type]runtime.Runtime{runtime.Kubernetes: fooRuntime}, nil
		}).Build()
		wo := &WatchOperation{opsmodels.Operation{RuntimeMap: map[models.Type]runtime.Runtime{runtime.Kubernetes: fooRuntime}}}
		err := wo.Watch(req)
		assert.Nil(t, err)
	})
}

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

func (f *fooWatchRuntime) Import(ctx context.Context, request *runtime.ImportRequest) *runtime.ImportResponse {
	return &runtime.ImportResponse{Resource: request.PlanResource}
}

func (f *fooWatchRuntime) Apply(ctx context.Context, request *runtime.ApplyRequest) *runtime.ApplyResponse {
	return nil
}

func (f *fooWatchRuntime) Read(ctx context.Context, request *runtime.ReadRequest) *runtime.ReadResponse {
	return nil
}

func (f *fooWatchRuntime) Delete(ctx context.Context, request *runtime.DeleteRequest) *runtime.DeleteResponse {
	return nil
}

func (f *fooWatchRuntime) Watch(ctx context.Context, request *runtime.WatchRequest) *runtime.WatchResponse {
	out := make(chan k8sWatch.Event)
	go func() {
		out <- k8sWatch.Event{
			Type:   k8sWatch.Deleted,
			Object: &unstructured.Unstructured{Object: barDeployment},
		}
		close(out)
	}()

	return &runtime.WatchResponse{
		Watchers: &runtime.SequentialWatchers{
			IDs:      []string{"apps/v1:Deployment:foo:bar"},
			Watchers: []<-chan k8sWatch.Event{out},
		},
		Status: nil,
	}
}

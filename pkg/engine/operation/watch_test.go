package operation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	k8sWatch "k8s.io/apimachinery/pkg/watch"

	"kusionstack.io/kusion/pkg/engine/models"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
)

func TestWatchOperation_Watch(t *testing.T) {
	req := &WatchRequest{
		Request: opsmodels.Request{
			Spec: &models.Spec{
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
	wo := &WatchOperation{Runtime: fooRuntime}
	err := wo.Watch(req)
	assert.Nil(t, err)
}

var fooRuntime = &fooWatchRuntime{}

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

type fooWatchRuntime struct{}

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
	out := make(chan runtime.RowData)
	go func() {
		i := 0
		for {
			switch i {
			case 0:
				out <- runtime.RowData{
					Event: k8sWatch.Event{
						Type: k8sWatch.Added,
					},
					Message: "readyReplicas: 0, desiredReplicas: 1, updatedReplicas: 0, availableReplicas: 0",
				}
				i++
			case 1:
				out <- runtime.RowData{
					Event: k8sWatch.Event{
						Type: k8sWatch.Modified,
					},
					Message: "readyReplicas: 1, desiredReplicas: 1, updatedReplicas: 1, availableReplicas: 1",
				}
				i++
			case 2:
				out <- runtime.RowData{
					Event: k8sWatch.Event{
						Type: k8sWatch.Deleted,
					},
					Message: "foo has been deleted",
				}
				i++
			default:
				close(out)
				return
			}
		}
	}()

	return &runtime.WatchResponse{
		MsgCh:  out,
		Status: nil,
	}
}

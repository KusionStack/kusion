//go:build ignore
// +build ignore

// ignore for test coverage
package kubernetes

import (
	"context"
	"github.com/bytedance/mockey"
	"os"
	"reflect"
	"testing"

	yamlv3 "gopkg.in/yaml.v3"

	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/models"
	jsonutil "kusionstack.io/kusion/pkg/util/json"
)

func TestKubernetesRuntime_Import(t *testing.T) {
	planServiceYaml, _ := os.ReadFile("testdata/plan_service.yaml")
	planSvc := &models.Resource{}
	yamlv3.Unmarshal(planServiceYaml, planSvc)

	lastAppliedYaml, _ := os.ReadFile("testdata/live_service_with_last_applied_annotation.yaml")
	lastAppliedObj := make(map[string]interface{})
	yamlv3.Unmarshal(lastAppliedYaml, lastAppliedObj)

	liveSvcYaml, _ := os.ReadFile("testdata/live_service.yaml")
	liveSvcObj := make(map[string]interface{})
	yamlv3.Unmarshal(liveSvcYaml, liveSvcObj)

	liveSvcImpYaml, _ := os.ReadFile("testdata/live_service_import_result.yaml")
	liveSvcImpObj := make(map[string]interface{})
	yamlv3.Unmarshal(liveSvcImpYaml, liveSvcImpObj)

	svcYaml, _ := os.ReadFile("testdata/service.yaml")
	svcObj := make(map[string]interface{})
	yamlv3.Unmarshal(svcYaml, svcObj)

	type args struct {
		ctx     context.Context
		request *runtime.ImportRequest
	}
	tests := []struct {
		name string
		args args
		want *runtime.ImportResponse
	}{
		{name: "import-svc-with-last-applied", args: struct {
			ctx     context.Context
			request *runtime.ImportRequest
		}{ctx: context.Background(), request: &runtime.ImportRequest{
			PlanResource: planSvc,
		}}, want: &runtime.ImportResponse{
			Resource: &models.Resource{
				ID:         planSvc.ResourceKey(),
				Type:       planSvc.Type,
				Attributes: svcObj,
				DependsOn:  planSvc.DependsOn,
				Extensions: planSvc.Extensions,
			},
			Status: nil,
		}},
		{name: "import-svc", args: args{
			ctx: context.Background(),
			request: &runtime.ImportRequest{
				PlanResource: planSvc,
			},
		}, want: &runtime.ImportResponse{
			Resource: &models.Resource{
				ID:         planSvc.ResourceKey(),
				Type:       planSvc.Type,
				Attributes: liveSvcImpObj,
				DependsOn:  planSvc.DependsOn,
				Extensions: planSvc.Extensions,
			},
			Status: nil,
		}},
	}

	mockey.PatchConvey(tests[0].name, t, func() {
		k := &KubernetesRuntime{}

		mockey.Mock(k.Read).To(func(
			k *KubernetesRuntime,
			ctx context.Context,
			request *runtime.ReadRequest,
		) *runtime.ReadResponse {
			return &runtime.ReadResponse{Resource: &models.Resource{
				ID:         planSvc.ResourceKey(),
				Type:       planSvc.Type,
				Attributes: lastAppliedObj,
				DependsOn:  planSvc.DependsOn,
				Extensions: planSvc.Extensions,
			}}
		}).Build()

		got := k.Import(tests[0].args.ctx, tests[0].args.request)
		if !reflect.DeepEqual(jsonutil.Marshal2PrettyString(got.Resource), jsonutil.Marshal2PrettyString(tests[0].want.Resource)) {
			t.Errorf("Import() = %v, want %v", jsonutil.Marshal2PrettyString(got.Resource), jsonutil.Marshal2PrettyString(tests[0].want.Resource))
		}
	})

	t.Run(tests[1].name, t, func() {
		k := &KubernetesRuntime{}
		mockey.Mock(k.Read).To(func(
			k *KubernetesRuntime,
			ctx context.Context,
			request *runtime.ReadRequest,
		) *runtime.ReadResponse {
			return &runtime.ReadResponse{Resource: &models.Resource{
				ID:         planSvc.ResourceKey(),
				Type:       planSvc.Type,
				Attributes: liveSvcObj,
				DependsOn:  planSvc.DependsOn,
				Extensions: planSvc.Extensions,
			}}
		}).Build()

		if got := k.Import(tests[1].args.ctx, tests[1].args.request); !reflect.DeepEqual(jsonutil.Marshal2PrettyString(got), jsonutil.Marshal2PrettyString(tests[1].want)) {
			t.Errorf("Import() = %v, want %v", jsonutil.Marshal2PrettyString(got), jsonutil.Marshal2PrettyString(tests[1].want))
		}
	})
}

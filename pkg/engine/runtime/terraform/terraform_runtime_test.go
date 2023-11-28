package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/stack"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform/tfops"
)

var testResource = intent.Resource{
	ID:   "hashicorp:local:local_file:kusion_example",
	Type: "Terraform",
	Attributes: map[string]interface{}{
		"content":  "kusion",
		"filename": "test.txt",
	},
	Extensions: map[string]interface{}{
		"provider":     "registry.terraform.io/hashicorp/local/2.2.3",
		"resourceType": "local_file",
	},
}

var fakeSR = &tfops.StateRepresentation{
	FormatVersion:    "0.2",
	TerraformVersion: "1.0.6",
	Values:           nil,
}

func TestTerraformRuntime(t *testing.T) {
	cwd, _ := os.Getwd()
	stack := &stack.Stack{
		Configuration: stack.Configuration{Name: "fakeStack"},
		Path:          filepath.Join(cwd, "fakePath"),
	}
	defer os.RemoveAll(stack.GetPath())
	tfRuntime := TerraformRuntime{
		WorkSpace: *tfops.NewWorkSpace(afero.Afero{Fs: afero.NewOsFs()}),
		mu:        &sync.Mutex{},
	}

	mockey.PatchConvey("ApplyDryRun", t, func() {
		mockApplySetup()
		data, err := os.ReadFile(filepath.Join("tfops", "test_data", "plan.out.json"))
		if err != nil {
			panic(err)
		}
		mockey.Mock((*tfops.WorkSpace).Plan).To(func(ws *tfops.WorkSpace, ctx context.Context) (*tfops.PlanRepresentation, error) {
			s := &tfops.PlanRepresentation{}
			if err = json.Unmarshal(data, s); err != nil {
				return nil, fmt.Errorf("json umarshal plan representation failed: %v", err)
			}
			return s, nil
		}).Build()
		response := tfRuntime.Apply(context.TODO(), &runtime.ApplyRequest{PlanResource: &testResource, DryRun: true, Stack: stack})
		assert.Equalf(t, nil, response.Status, "Execute(%v)", "Apply")
	})
	mockey.PatchConvey("Apply", t, func() {
		mockApplySetup()
		response := tfRuntime.Apply(context.TODO(), &runtime.ApplyRequest{PlanResource: &testResource, DryRun: false, Stack: stack})
		assert.Equalf(t, nil, response.Status, "Execute(%v)", "Apply")
	})

	mockey.PatchConvey("Read", t, func() {
		mockey.Mock((*tfops.WorkSpace).InitWorkSpace).To(func(ws *tfops.WorkSpace, ctx context.Context) error {
			return nil
		}).Build()
		response := tfRuntime.Read(context.TODO(), &runtime.ReadRequest{PlanResource: &testResource, Stack: stack})
		assert.Equalf(t, nil, response.Status, "Execute(%v)", "Read")
	})

	mockey.PatchConvey("Delete", t, func() {
		mockey.Mock((*tfops.WorkSpace).InitWorkSpace).To(func(ws *tfops.WorkSpace, ctx context.Context) error {
			return nil
		}).Build()
		// mock destroy
		mockey.Mock((*tfops.WorkSpace).Destroy).To(func(ws *tfops.WorkSpace, ctx context.Context) error {
			return nil
		}).Build()
		response := tfRuntime.Delete(context.TODO(), &runtime.DeleteRequest{Resource: &testResource, Stack: stack})
		assert.Equalf(t, nil, response.Status, "Execute(%v)", "Delete")
	})
}

func mockApplySetup() {
	mockey.Mock((*tfops.WorkSpace).InitWorkSpace).To(func(ws *tfops.WorkSpace, ctx context.Context) error {
		return nil
	}).Build()
	mockey.Mock((*tfops.WorkSpace).Apply).To(func(ws *tfops.WorkSpace, ctx context.Context) (*tfops.StateRepresentation, error) {
		return fakeSR, nil
	}).Build()
	mockey.Mock((*tfops.WorkSpace).GetProvider).To(func(ws *tfops.WorkSpace) (string, error) {
		return "registry.terraform.io/hashicorp/local/2.2.3", nil
	}).Build()
}

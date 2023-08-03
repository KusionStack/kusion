package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"bou.ke/monkey"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform/tfops"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/projectstack"
)

var testResource = models.Resource{
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
	stack := &projectstack.Stack{
		StackConfiguration: projectstack.StackConfiguration{Name: "fakeStack"},
		Path:               filepath.Join(cwd, "fakePath"),
	}
	defer os.RemoveAll(stack.GetPath())
	tfRuntime := TerraformRuntime{
		WorkSpace: *tfops.NewWorkSpace(afero.Afero{Fs: afero.NewOsFs()}),
		mu:        &sync.Mutex{},
	}

	t.Run("ApplyDryRun", func(t *testing.T) {
		defer monkey.UnpatchAll()

		mockApplySetup()

		data, err := os.ReadFile(filepath.Join("tfops", "test_data", "plan.out.json"))
		if err != nil {
			panic(err)
		}
		monkey.Patch((*tfops.WorkSpace).Plan, func(ws *tfops.WorkSpace, ctx context.Context) (*tfops.PlanRepresentation, error) {
			s := &tfops.PlanRepresentation{}
			if err = json.Unmarshal(data, s); err != nil {
				return nil, fmt.Errorf("json umarshal plan representation failed: %v", err)
			}
			return s, nil
		})

		response := tfRuntime.Apply(context.TODO(), &runtime.ApplyRequest{PlanResource: &testResource, DryRun: true, Stack: stack})
		assert.Equalf(t, nil, response.Status, "Execute(%v)", "Apply")
	})
	t.Run("Apply", func(t *testing.T) {
		defer monkey.UnpatchAll()

		mockApplySetup()

		response := tfRuntime.Apply(context.TODO(), &runtime.ApplyRequest{PlanResource: &testResource, DryRun: false, Stack: stack})
		assert.Equalf(t, nil, response.Status, "Execute(%v)", "Apply")
	})

	t.Run("Read", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch((*tfops.WorkSpace).InitWorkSpace, func(ws *tfops.WorkSpace, ctx context.Context) error {
			return nil
		})

		response := tfRuntime.Read(context.TODO(), &runtime.ReadRequest{PlanResource: &testResource, Stack: stack})
		assert.Equalf(t, nil, response.Status, "Execute(%v)", "Read")
	})

	t.Run("Delete", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch((*tfops.WorkSpace).InitWorkSpace, func(ws *tfops.WorkSpace, ctx context.Context) error {
			return nil
		})
		// mock destroy
		monkey.Patch((*tfops.WorkSpace).Destroy, func(ws *tfops.WorkSpace, ctx context.Context) error {
			return nil
		})

		response := tfRuntime.Delete(context.TODO(), &runtime.DeleteRequest{Resource: &testResource, Stack: stack})
		assert.Equalf(t, nil, response.Status, "Execute(%v)", "Delete")
	})
}

func mockApplySetup() {
	monkey.Patch((*tfops.WorkSpace).InitWorkSpace, func(ws *tfops.WorkSpace, ctx context.Context) error {
		return nil
	})
	monkey.Patch((*tfops.WorkSpace).Apply, func(ws *tfops.WorkSpace, ctx context.Context) (*tfops.StateRepresentation, error) {
		return fakeSR, nil
	})
	monkey.Patch((*tfops.WorkSpace).GetProvider, func(ws *tfops.WorkSpace) (string, error) {
		return "registry.terraform.io/hashicorp/local/2.2.3", nil
	})
}

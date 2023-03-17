package terraform

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform/tfops"
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
		response := tfRuntime.Apply(context.TODO(), &runtime.ApplyRequest{PlanResource: &testResource, DryRun: true, Stack: stack})
		assert.Equalf(t, nil, response.Status, "Execute(%v)", "Apply")
	})
	t.Run("Apply", func(t *testing.T) {
		response := tfRuntime.Apply(context.TODO(), &runtime.ApplyRequest{PlanResource: &testResource, DryRun: false, Stack: stack})
		assert.Equalf(t, nil, response.Status, "Execute(%v)", "Apply")
	})

	t.Run("Read", func(t *testing.T) {
		response := tfRuntime.Read(context.TODO(), &runtime.ReadRequest{PlanResource: &testResource, Stack: stack})
		assert.Equalf(t, nil, response.Status, "Execute(%v)", "Read")
	})

	t.Run("Delete", func(t *testing.T) {
		response := tfRuntime.Delete(context.TODO(), &runtime.DeleteRequest{Resource: &testResource, Stack: stack})
		assert.Equalf(t, nil, response.Status, "Execute(%v)", "Delete")
	})
}

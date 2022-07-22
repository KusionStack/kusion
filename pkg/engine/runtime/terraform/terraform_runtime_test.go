package terraform

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform/tfops"
)

var testResource = models.Resource{
	ID:   "example",
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
	wd, _ := tfops.GetWorkSpaceDir()
	defer os.RemoveAll(filepath.Join(wd, testResource.ID))
	tfRuntime, _ := NewTerraformRuntime()
	t.Run("ApplyDryRun", func(t *testing.T) {
		response := tfRuntime.Apply(context.TODO(), &runtime.ApplyRequest{PlanResource: &testResource, DryRun: true})
		assert.Equalf(t, nil, response.Status, "Execute(%v)", "Apply")
	})
	t.Run("Apply", func(t *testing.T) {
		response := tfRuntime.Apply(context.TODO(), &runtime.ApplyRequest{PlanResource: &testResource, DryRun: false})
		assert.Equalf(t, nil, response.Status, "Execute(%v)", "Apply")
	})

	t.Run("Read", func(t *testing.T) {
		response := tfRuntime.Read(context.TODO(), &runtime.ReadRequest{PlanResource: &testResource})
		assert.Equalf(t, nil, response.Status, "Execute(%v)", "Read")
	})

	t.Run("Delete", func(t *testing.T) {
		response := tfRuntime.Delete(context.TODO(), &runtime.DeleteRequest{Resource: &testResource})
		assert.Equalf(t, nil, response.Status, "Execute(%v)", "Delete")
	})
}

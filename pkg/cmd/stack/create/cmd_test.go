package create

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	projutil "kusionstack.io/kusion/pkg/cmd/project/util"
	"kusionstack.io/kusion/pkg/cmd/stack/util"
	"kusionstack.io/kusion/pkg/modules/generators/workload/secret"
)

func TestNewCmd(t *testing.T) {
	// Create a temporary project and stack directory for unit tests.
	randomSuffix := secret.GenerateRandomString(16)
	tmpTestRootDir, err := os.MkdirTemp("", "kusion-test-stack-util-"+randomSuffix)
	if err != nil {
		t.Fatalf("failed to create temporary test root directory: %v", err)
	}
	defer os.RemoveAll(tmpTestRootDir)

	tmpProjectDir, err := os.MkdirTemp(tmpTestRootDir, "my-project")
	if err != nil {
		t.Fatalf("failed to create temporary project directory: %v", err)
	}
	projectYAMLPath := filepath.Join(tmpProjectDir, projutil.ProjectYAMLFile)
	projectYAMLContent := fmt.Sprintf(projutil.ProjectYAMLTemplate, "my-project")
	if err = os.WriteFile(projectYAMLPath, []byte(projectYAMLContent), 0o644); err != nil {
		t.Fatalf("failed to create project.yaml file in project directory '%s': %v", tmpProjectDir, err)
	}

	tmpRefStackDir, err := os.MkdirTemp(tmpProjectDir, "dev")
	if err != nil {
		t.Fatalf("failed to create temporary stack directory: %v", err)
	}
	stackYAMLPath := filepath.Join(tmpRefStackDir, util.StackYAMLFile)
	stackYAMLContent := fmt.Sprintf(util.StackYAMLTemplate, "dev")
	if err = os.WriteFile(stackYAMLPath, []byte(stackYAMLContent), 0o644); err != nil {
		t.Fatalf("failed to create stack.yaml file in stack directory '%s': %v", tmpRefStackDir, err)
	}

	kclModPath := filepath.Join(tmpRefStackDir, util.KCLModFile)
	kclModContent := fmt.Sprint(util.KCLModFileTemplate)
	if err = os.WriteFile(kclModPath, []byte(kclModContent), 0o644); err != nil {
		t.Fatalf("failed to create kcl.mod file in stack directory '%s': %v", tmpRefStackDir, err)
	}

	mainKCLPath := filepath.Join(tmpRefStackDir, util.MainKCLFile)
	mainKCLContent := fmt.Sprint(util.MainKCLFileTemplate)
	if err = os.WriteFile(mainKCLPath, []byte(mainKCLContent), 0o644); err != nil {
		t.Fatalf("failed to create main.k file in stack directory '%s': %v", tmpRefStackDir, err)
	}

	mockey.PatchConvey("mock options.Complete", t, func() {
		mockey.Mock((*Options).Complete).To(func(o *Options, args []string) error {
			o.Name = "prod"
			o.StackDir = filepath.Join(tmpProjectDir, "prod")
			o.Flags.ProjectDir = tmpProjectDir
			o.Flags.CopyFrom = tmpRefStackDir

			return nil
		}).Build()

		cmd := NewCmd()
		err := cmd.Execute()
		assert.Nil(t, err)
	})
}

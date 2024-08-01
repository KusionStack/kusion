package create

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	projutil "kusionstack.io/kusion/pkg/cmd/project/util"
	"kusionstack.io/kusion/pkg/cmd/stack/util"
)

func TestOptions_Complete(t *testing.T) {
	t.Run("not one arg", func(t *testing.T) {
		opts := NewOptions()
		args := []string{"dev", "pre", "prod"}

		err := opts.Complete(args)
		assert.ErrorContains(t, err, ErrNotOneArg.Error())
	})

	t.Run("target project not specified", func(t *testing.T) {
		mockey.PatchConvey("mock os.Getwd", t, func() {
			mockey.Mock(os.Getwd).To(func() (dir string, err error) {
				return "/dir/to/my-project", nil
			}).Build()

			opts := NewOptions()
			args := []string{"dev"}

			err := opts.Complete(args)
			assert.NoError(t, err)
			assert.Equal(t, "/dir/to/my-project", opts.ProjectDir)
			assert.Equal(t, "/dir/to/my-project/dev", opts.StackDir)
			assert.Equal(t, "", opts.CopyFrom)
		})
	})

	t.Run("specified target project directory", func(t *testing.T) {
		mockey.PatchConvey("mock filepath.Abs", t, func() {
			mockey.Mock(filepath.Abs).To(func(_ string) (string, error) {
				return "/dir/to/my-project", nil
			}).Build()

			opts := NewOptions()
			opts.ProjectDir = "/dir/to/my-project"
			args := []string{"dev"}

			err := opts.Complete(args)
			assert.NoError(t, err)
			assert.Equal(t, "/dir/to/my-project/dev", opts.StackDir)
			assert.Equal(t, "", opts.CopyFrom)
		})
	})

	t.Run("specified referenced stack directory", func(t *testing.T) {
		mockey.PatchConvey("mock os.Getwd", t, func() {
			mockey.Mock(os.Getwd).To(func() (dir string, err error) {
				return "/dir/to/my-project", nil
			}).Build()

			opts := NewOptions()
			opts.CopyFrom = "dev"
			args := []string{"prod"}

			err := opts.Complete(args)
			assert.NoError(t, err)
			assert.Equal(t, "/dir/to/my-project", opts.ProjectDir)
			assert.Equal(t, "/dir/to/my-project/prod", opts.StackDir)
			assert.Equal(t, "/dir/to/my-project/dev", opts.CopyFrom)
		})
	})
}

func TestOptions_Validate(t *testing.T) {
	t.Run("failed to validate stack name", func(t *testing.T) {
		mockey.PatchConvey("mock validate stack name", t, func() {
			mockey.Mock(util.ValidateStackName).
				Return(errors.New("failed to validate stack name")).Build()

			opts := NewOptions()
			err := opts.Validate()
			assert.ErrorContains(t, err, "failed to validate stack name")
		})
	})

	t.Run("failed to validate project directory", func(t *testing.T) {
		mockey.PatchConvey("mock validate stack name and project directory", t, func() {
			mockey.Mock(util.ValidateStackName).Return(nil).Build()
			mockey.Mock(util.ValidateProjectDir).
				Return(errors.New("failed to validate project directory")).Build()

			opts := NewOptions()
			err := opts.Validate()
			assert.ErrorContains(t, err, "failed to validate project directory")
		})
	})

	t.Run("failed to validate stack directory", func(t *testing.T) {
		mockey.PatchConvey("mock validate stack name, project directory and stack directory", t, func() {
			mockey.Mock(util.ValidateStackName).Return(nil).Build()
			mockey.Mock(util.ValidateProjectDir).Return(nil).Build()
			mockey.Mock(util.ValidateStackDir).
				Return(errors.New("failed to validate stack directory")).Build()

			opts := NewOptions()
			err := opts.Validate()
			assert.ErrorContains(t, err, "failed to validate stack directory")
		})
	})

	t.Run("failed to validate referenced stack directory", func(t *testing.T) {
		mockey.PatchConvey("mock validate stack name, project directory and stack directory", t, func() {
			mockey.Mock(util.ValidateStackName).Return(nil).Build()
			mockey.Mock(util.ValidateProjectDir).Return(nil).Build()
			mockey.Mock(util.ValidateStackDir).Return(nil).Build()
			mockey.Mock(util.ValidateRefStackDir).
				Return(errors.New("failed to validate referenced stack directory")).Build()

			opts := NewOptions()
			opts.CopyFrom = "/dir/to/referenced-stack"
			err := opts.Validate()
			assert.ErrorContains(t, err, "failed to validate referenced stack directory")
		})
	})

	t.Run("successfully validate options", func(t *testing.T) {
		mockey.PatchConvey("mock validate stack name, project directory and stack directory", t, func() {
			mockey.Mock(util.ValidateStackName).Return(nil).Build()
			mockey.Mock(util.ValidateProjectDir).Return(nil).Build()
			mockey.Mock(util.ValidateStackDir).Return(nil).Build()
			mockey.Mock(util.ValidateRefStackDir).Return(nil).Build()

			opts := NewOptions()
			err := opts.Validate()
			assert.NoError(t, err)
		})
	})
}

func TestOptions_Run(t *testing.T) {
	// Create a temporary project and stack directory for unit tests.
	suffix := "options-test"
	tmpTestRootDir, err := os.MkdirTemp("", "kusion-test-stack-util-"+suffix)
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

	t.Run("referenced stack directory not specified", func(t *testing.T) {
		opts := NewOptions()
		opts.Name = "my-dev"
		opts.StackDir = filepath.Join(tmpProjectDir, "my-dev")
		opts.ProjectDir = tmpProjectDir

		_ = opts.Validate()
		err := opts.Run()
		assert.NoError(t, err)
	})

	t.Run("specified referenced stack directory", func(t *testing.T) {
		opts := NewOptions()
		opts.Name = "prod"
		opts.StackDir = filepath.Join(tmpProjectDir, "prod")
		opts.ProjectDir = tmpProjectDir
		opts.CopyFrom = tmpRefStackDir

		_ = opts.Validate()
		err := opts.Run()
		assert.NoError(t, err)
	})
}

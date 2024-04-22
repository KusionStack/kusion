package create

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"kusionstack.io/kusion/pkg/modules/generators/workload/secret"
)

func TestNewCmd(t *testing.T) {
	t.Run("successfully create project", func(t *testing.T) {
		// Create a temporary project directory for unit test.
		randomSuffix := secret.GenerateRandomString(16)
		tmpTestRootDir, err := os.MkdirTemp("", "kusion-test-project-create-"+randomSuffix)
		if err != nil {
			t.Fatalf("failed to create temporary test root directory: %v", err)
		}
		defer os.RemoveAll(tmpTestRootDir)

		mockey.PatchConvey("mock options.Complete", t, func() {
			mockey.Mock((*Options).Complete).To(func(o *Options, args []string) error {
				o.Name = filepath.Base(tmpTestRootDir)
				o.Flags.ProjectDir = tmpTestRootDir

				return nil
			}).Build()

			cmd := NewCmd()
			err := cmd.Execute()
			assert.Nil(t, err)
		})
	})
}

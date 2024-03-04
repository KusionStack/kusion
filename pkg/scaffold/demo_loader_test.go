package scaffold

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/workspace"
)

func TestGenDemoProject(t *testing.T) {
	// Create temporary project directory for unit test.
	tmpProjectDir, err := os.MkdirTemp("", "kusion-quickstart-test")
	if err != nil {
		t.Fatalf("failed to create temporary project directory: %v", err)
	}
	defer os.RemoveAll(tmpProjectDir)

	t.Run("failed to create default workspace", func(t *testing.T) {
		mockey.PatchConvey("mock workspace related functions", t, func() {
			mockey.Mock(workspace.GetWorkspaceByDefaultOperator).
				To(func(name string) (*v1.Workspace, error) {
					return &v1.Workspace{}, workspace.ErrWorkspaceNotExist
				}).Build()
			mockey.Mock(workspace.CreateWorkspaceByDefaultOperator).
				To(func(ws *v1.Workspace) error {
					return errors.New("failed to create default workspace")
				}).Build()

			err := GenDemoProject("/dir/to/quickstart", "quickstart")
			assert.ErrorContains(t, err, "failed to create default workspace")
		})
	})

	t.Run("failed to create destination directory or file path", func(t *testing.T) {
		mockey.PatchConvey("mock workspace related function", t, func() {
			mockey.Mock(workspace.GetWorkspaceByDefaultOperator).Return(nil, nil).Build()

			err := GenDemoProject("/dir/to/kusion-project/not-exists", "not-exists")
			assert.NotNil(t, err)
		})
	})

	t.Run("successfully creates the demo project", func(t *testing.T) {
		err := GenDemoProject(tmpProjectDir, filepath.Base(tmpProjectDir))
		assert.Nil(t, err)
	})
}

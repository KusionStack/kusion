package scaffold

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/backend"
	workspacestorages "kusionstack.io/kusion/pkg/workspace/storages"
)

func TestGenDemoProject(t *testing.T) {
	// Create temporary project directory for unit test.
	tmpProjectDir, err := os.MkdirTemp("", "kusion-quickstart-test")
	if err != nil {
		t.Fatalf("failed to create temporary project directory: %v", err)
	}
	defer os.RemoveAll(tmpProjectDir)

	t.Run("failed to init default workspace", func(t *testing.T) {
		mockey.PatchConvey("mock workspace related functions", t, func() {
			mockey.Mock(backend.NewWorkspaceStorage).Return(nil, errors.New("failed to create default workspace")).Build()

			err = GenDemoProject("/dir/to/quickstart", "quickstart")
			assert.ErrorContains(t, err, "failed to create default workspace")
		})
	})

	t.Run("failed to create destination directory or file path", func(t *testing.T) {
		mockey.PatchConvey("mock workspace related function", t, func() {
			mockey.Mock(backend.NewWorkspaceStorage).Return(&workspacestorages.LocalStorage{}, nil).Build()

			err = GenDemoProject("/dir/to/kusion-project/not-exists", "not-exists")
			assert.NotNil(t, err)
		})
	})

	t.Run("successfully creates the demo project", func(t *testing.T) {
		err = GenDemoProject(tmpProjectDir, filepath.Base(tmpProjectDir))
		assert.Nil(t, err)
	})
}

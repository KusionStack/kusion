package version

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/version"
)

func TestNewCmdVersion(t *testing.T) {
	_, err := version.NewInfo()
	assert.Nil(t, err)

	t.Run("ExportJSON", func(t *testing.T) {
		cmd := NewCmdVersion()
		err := cmd.Flags().Set("json", "true")
		assert.Nil(t, err)
		err = cmd.Execute()
		assert.Nil(t, err)
	})

	t.Run("ExportYaml", func(t *testing.T) {
		cmd := NewCmdVersion()
		err := cmd.Flags().Set("yaml", "true")
		assert.Nil(t, err)
		err = cmd.Execute()
		assert.Nil(t, err)
	})

	t.Run("ShortString", func(t *testing.T) {
		cmd := NewCmdVersion()
		err := cmd.Flags().Set("short", "true")
		assert.Nil(t, err)
		err = cmd.Execute()
		assert.Nil(t, err)
	})

	t.Run("default", func(t *testing.T) {
		cmd := NewCmdVersion()
		err := cmd.Execute()
		assert.Nil(t, err)
	})
}

package version

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/version"
)

func TestNewCmdVersion(t *testing.T) {
	_, err := version.NewInfo()
	assert.Nil(t, err)

	t.Run("json", func(t *testing.T) {
		cmd := NewCmdVersion()
		err := cmd.Flags().Set("output", "json")
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

package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCmdEnv(t *testing.T) {
	t.Run("json", func(t *testing.T) {
		cmd := NewCmdEnv()
		err := cmd.Flags().Set("json", "true")
		assert.Nil(t, err)
		err = cmd.Execute()
		assert.Nil(t, err)
	})

	t.Run("default", func(t *testing.T) {
		cmd := NewCmdEnv()
		err := cmd.Execute()
		assert.Nil(t, err)
	})
}

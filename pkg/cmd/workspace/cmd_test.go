package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCmd(t *testing.T) {
	t.Run("successfully get workspace help", func(t *testing.T) {
		cmd := NewCmd()
		assert.NotNil(t, cmd)
	})
}

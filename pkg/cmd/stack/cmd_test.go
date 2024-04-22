package stack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCmd(t *testing.T) {
	t.Run("successfully get stack help", func(t *testing.T) {
		cmd := NewCmd()
		assert.NotNil(t, cmd)
	})
}

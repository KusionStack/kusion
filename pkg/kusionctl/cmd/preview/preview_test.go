package preview

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCmdPreview(t *testing.T) {
	t.Run("validate error", func(t *testing.T) {
		cmd := NewCmdPreview()
		err := cmd.Execute()
		assert.NotNil(t, err)
	})
}

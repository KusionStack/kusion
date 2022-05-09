package destroy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDestroyCommandRun(t *testing.T) {
	t.Run("validate error", func(t *testing.T) {
		cmd := NewCmdDestroy()
		err := cmd.Execute()
		assert.NotNil(t, err)
	})
}

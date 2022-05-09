package apply

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyCommandRun(t *testing.T) {
	t.Run("validate error", func(t *testing.T) {
		cmd := NewCmdApply()
		err := cmd.Execute()
		assert.NotNil(t, err)
	})
}

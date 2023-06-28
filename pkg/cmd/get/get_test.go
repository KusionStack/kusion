package get

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCommandRun(t *testing.T) {
	// TODO: fix#356 better unit test
	t.Run("validate error", func(t *testing.T) {
		cmd := NewCmdGet()
		err := cmd.Execute()
		assert.NotNil(t, err)
	})
}

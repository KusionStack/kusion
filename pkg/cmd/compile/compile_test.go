package compile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCmdCompile(t *testing.T) {
	t.Run("compile", func(t *testing.T) {
		cmd := NewCmdCompile()
		err := cmd.Execute()
		assert.Errorf(t, err, "this command is deprecated. Please use `kusion build` to generate the Intent instead")
	})
}

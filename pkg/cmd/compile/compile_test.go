package compile

import (
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdCompile(t *testing.T) {
	defer monkey.UnpatchAll()

	monkey.Patch((*CompileOptions).Complete, func(o *CompileOptions, args []string) {
		o.Output = "stdout"
	})
	monkey.Patch((*CompileOptions).Run, func(*CompileOptions) error {
		return nil
	})

	t.Run("compile success", func(t *testing.T) {
		cmd := NewCmdCompile()
		err := cmd.Execute()
		assert.Nil(t, err)
	})
}

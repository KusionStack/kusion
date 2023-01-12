package check

import (
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/kusionctl/cmd/compile"
)

func TestNewCmdCheck(t *testing.T) {
	t.Run("", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch((*compile.CompileOptions).Complete, func(o *compile.CompileOptions, args []string) {
			o.Output = "stdout"
		})
		monkey.Patch((*compile.CompileOptions).Run, func(*compile.CompileOptions) error {
			return nil
		})

		cmd := NewCmdCheck()
		err := cmd.Execute()
		assert.Nil(t, err)
	})
}

package check

import (
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/cmd/compile"
)

func TestNewCmdCheck(t *testing.T) {
	t.Run("", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch((*compile.Options).Complete, func(o *compile.Options, args []string) error {
			o.Output = "stdout"
			return nil
		})
		monkey.Patch((*compile.Options).Run, func(*compile.Options) error {
			return nil
		})

		cmd := NewCmdCheck()
		err := cmd.Execute()
		assert.Nil(t, err)
	})
}

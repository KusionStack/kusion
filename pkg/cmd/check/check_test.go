package check

import (
	"github.com/bytedance/mockey"
	"testing"

	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/cmd/compile"
)

func TestNewCmdCheck(t *testing.T) {
	mockey.PatchConvey("", t, func() {

		mockey.Mock((*compile.CompileOptions).Complete).To(func(o *compile.CompileOptions, args []string) error {
			o.Output = "stdout"
			return nil
		}).Build()
		mockey.Mock((*compile.CompileOptions).Run).To(func(*compile.CompileOptions) error {
			return nil
		}).Build()

		cmd := NewCmdCheck()
		err := cmd.Execute()
		assert.Nil(t, err)
	})
}

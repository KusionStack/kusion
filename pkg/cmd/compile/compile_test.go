package compile

import (
	"github.com/bytedance/mockey"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCmdCompile(t *testing.T) {
	mockey.PatchConvey("compile success", t, func() {
		mockey.Mock((*CompileOptions).Complete).To(func(o *CompileOptions, args []string) {
			o.Output = "stdout"
		}).Build()
		mockey.Mock((*CompileOptions).Run).To(func(*CompileOptions) error {
			return nil
		}).Build()
		cmd := NewCmdCompile()
		err := cmd.Execute()
		assert.Nil(t, err)
	})
}

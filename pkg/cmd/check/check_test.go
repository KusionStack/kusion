package check

import (
	"github.com/bytedance/mockey"
	"testing"

	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/cmd/compile"
)

func TestNewCmdCheck(t *testing.T) {
	mockey.PatchConvey("", t, func() {
		mockey.Mock((*compile.Options).Complete).To(func(o *compile.Options, args []string) error {
			o.Output = "stdout"
			return nil
		}).Build()
		mockey.Mock((*compile.Options).Run).To(func(*compile.Options) error {
			return nil
		}).Build()
		cmd := NewCmdCheck()
		err := cmd.Execute()
		assert.Nil(t, err)
	})
}

package cmdswitch

import (
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
)

func TestNewCmd(t *testing.T) {
	t.Run("successfully switch workspace", func(t *testing.T) {
		mockey.PatchConvey("mock cmd", t, func() {
			mockey.Mock((*Options).Complete).To(func(o *Options, args []string) error {
				o.Name = "dev"
				return nil
			}).Build()
			mockey.Mock((*Options).Run).Return(nil).Build()

			cmd := NewCmd()
			err := cmd.Execute()
			assert.Nil(t, err)
		})
	})
}

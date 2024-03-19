package list

import (
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
)

func TestNewCmd(t *testing.T) {
	t.Run("successfully list workspace", func(t *testing.T) {
		mockey.PatchConvey("mock cmd", t, func() {
			mockey.Mock((*Options).Run).Return(nil).Build()

			cmd := NewCmd()
			err := cmd.Execute()
			assert.Nil(t, err)
		})
	})
}

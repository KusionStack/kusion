package init

import (
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
)

func TestNewCmd(t *testing.T) {
	t.Run("successfully initiate a demo project", func(t *testing.T) {
		mockey.PatchConvey("mock complete, validate and run", t, func() {
			mockey.Mock((*Options).Complete).Return(nil).Build()
			mockey.Mock((*Options).Validate).Return(nil).Build()
			mockey.Mock((*Options).Run).Return(nil).Build()

			cmd := NewCmd()
			err := cmd.Execute()
			assert.Nil(t, err)
		})
	})
}

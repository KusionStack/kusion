package create

import (
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
)

func TestNewCmd(t *testing.T) {
	t.Run("failed to create empty project", func(t *testing.T) {
		mockey.PatchConvey("mock complete", t, func() {
			mockey.Mock((*Options).Complete).Return(nil).Build()

			cmd := NewCmd()
			err := cmd.Execute()
			assert.ErrorContains(t, err, "the project name must not be empty")
		})
	})

	t.Run("successfully create a project", func(t *testing.T) {
		mockey.PatchConvey("mock complete and run", t, func() {
			mockey.Mock((*Options).Complete).To(func(o *Options, args []string) error {
				o.Name = "new-project"
				return nil
			}).Build()
			mockey.Mock((*Options).Run).Return(nil).Build()

			cmd := NewCmd()
			err := cmd.Execute()
			assert.Nil(t, err)
		})
	})
}

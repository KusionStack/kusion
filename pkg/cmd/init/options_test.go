package init

import (
	"errors"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"kusionstack.io/kusion/pkg/cmd/init/util"
	"kusionstack.io/kusion/pkg/scaffold"
)

func TestOptions_Complete(t *testing.T) {
	t.Run("not empty args", func(t *testing.T) {
		opts := NewOptions()
		args := []string{"quickstart"}

		err := opts.Complete(args)
		assert.ErrorContains(t, err, ErrNotEmptyArgs.Error())
	})

	t.Run("failed to get directory and name", func(t *testing.T) {
		mockey.PatchConvey("mock util.GetDirAndName", t, func() {
			mockey.Mock(util.GetDirAndName).To(func() (dir, name string, err error) {
				return "", "", errors.New("failed to get directory and name")
			}).Build()

			opts := NewOptions()
			args := []string{}

			err := opts.Complete(args)
			assert.ErrorContains(t, err, "failed to get directory and name")
		})
	})

	t.Run("successfully complete the options", func(t *testing.T) {
		mockey.PatchConvey("mock util.GetDirAndName", t, func() {
			mockey.Mock(util.GetDirAndName).To(func() (dir, name string, err error) {
				return "/dir/to/quickstart", "quickstart", nil
			}).Build()

			opts := NewOptions()
			opts.Flags.ProjectDir = "/dir/to/my-project"
			args := []string{}

			err := opts.Complete(args)
			assert.Nil(t, err)
			assert.Equal(t, "/dir/to/my-project", opts.ProjectDir)
		})
	})
}

func TestOptions_Validate(t *testing.T) {
	t.Run("failed to validate project directory", func(t *testing.T) {
		mockey.PatchConvey("mock util.ValidateProjectDir", t, func() {
			mockey.Mock(util.ValidateProjectDir).
				Return(errors.New("failed to validate project directory")).Build()

			opts := NewOptions()
			err := opts.Validate()
			assert.ErrorContains(t, err, "failed to validate project directory")
		})
	})

	t.Run("failed to validate project name", func(t *testing.T) {
		mockey.PatchConvey("mock util.ValidateProjectDir and util.ValidateProjectName", t, func() {
			mockey.Mock(util.ValidateProjectDir).Return(nil).Build()
			mockey.Mock(util.ValidateProjectName).
				Return(errors.New("failed to validate project name")).Build()

			opts := NewOptions()
			err := opts.Validate()
			assert.ErrorContains(t, err, "failed to validate project name")
		})
	})

	t.Run("successfully validate the options", func(t *testing.T) {
		mockey.PatchConvey("mock util.ValidateProjectDir and util.ValidateProjectName", t, func() {
			mockey.Mock(util.ValidateProjectDir).Return(nil).Build()
			mockey.Mock(util.ValidateProjectName).Return(nil).Build()

			opts := NewOptions()
			err := opts.Validate()
			assert.Nil(t, err)
		})
	})
}

func TestOptions_Run(t *testing.T) {
	t.Run("failed to generate demo project", func(t *testing.T) {
		mockey.PatchConvey("mock scaffold.GenDemoProject", t, func() {
			mockey.Mock(scaffold.GenDemoProject).
				Return(errors.New("failed to generate demo project")).Build()

			opts := NewOptions()
			err := opts.Run()
			assert.ErrorContains(t, err, "failed to generate demo project")
		})
	})

	t.Run("successfully initiate the demo project", func(t *testing.T) {
		mockey.PatchConvey("mock scaffold.GenDemoProject", t, func() {
			mockey.Mock(scaffold.GenDemoProject).Return(nil).Build()

			opts := NewOptions()
			err := opts.Run()
			assert.Nil(t, err)
		})
	})
}

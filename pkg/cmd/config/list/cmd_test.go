package list

import (
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/internal.kusion.io/v1"
	"kusionstack.io/kusion/pkg/config"
)

func TestNewCmd(t *testing.T) {
	t.Run("successfully list configs", func(t *testing.T) {
		mockey.PatchConvey("mock cmd", t, func() {
			mockey.Mock(Run).Return(nil).Build()

			cmd := NewCmd()
			err := cmd.Execute()
			assert.Nil(t, err)
		})
	})
}

func TestValidate(t *testing.T) {
	testcases := []struct {
		name    string
		args    []string
		success bool
	}{
		{
			name:    "valid args",
			args:    nil,
			success: true,
		},
		{
			name:    "invalid args not empty",
			args:    []string{"invalid"},
			success: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.args)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestRun(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
	}{
		{
			name:    "successfully run",
			success: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock get config", t, func() {
				mockey.Mock(config.GetConfig).Return(&v1.Config{}, nil).Build()

				err := Run()
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

//go:build !arm64
// +build !arm64

package init

import (
	"github.com/bytedance/mockey"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CmdInit(t *testing.T) {
	mockey.PatchConvey("cmd init", t, func() {
		// patch human interact
		patchChooseTemplate()
		patchPromptValue()

		cmd := NewCmdInit()
		_ = cmd.Flags().Set("project-name", "test")
		// clean data
		defer os.RemoveAll("test")

		err := cmd.Execute()
		assert.Nil(t, err)
	})
}

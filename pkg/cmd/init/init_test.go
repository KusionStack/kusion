//go:build !arm64
// +build !arm64

package init

import (
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

func Test_CmdInit(t *testing.T) {
	// patch human interact
	patchChooseTemplate()
	patchPromptValue()
	defer monkey.UnpatchAll()

	cmd := NewCmdInit()
	_ = cmd.Flags().Set("project-name", "test")
	// clean data
	defer os.RemoveAll("test")

	err := cmd.Execute()
	assert.Nil(t, err)
}

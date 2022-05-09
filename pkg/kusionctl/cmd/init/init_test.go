package init

import (
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

func Test_CmdInit(t *testing.T) {
	// patch human interact
	patch_chooseTemplate()
	patch_promptValue()
	defer monkey.UnpatchAll()

	cmd := NewCmdInit()
	cmd.Flags().Set("project-name", "test")
	// clean data
	defer os.RemoveAll("test")

	err := cmd.Execute()
	assert.Nil(t, err)
}

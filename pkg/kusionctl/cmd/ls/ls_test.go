package ls

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLsCommandRun(t *testing.T) {
	cmd := NewCmdLs()
	cmd.SetArgs([]string{project.Path})
	cmd.Flags().Set("format", "tree")
	err := cmd.Execute()
	assert.Nil(t, err)
}

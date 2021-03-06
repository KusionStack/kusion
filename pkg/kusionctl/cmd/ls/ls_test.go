//go:build !arm64
// +build !arm64

package ls

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLsCommandRun(t *testing.T) {
	cmd := NewCmdLs()
	cmd.SetArgs([]string{project.Path})
	_ = cmd.Flags().Set("format", "tree")
	err := cmd.Execute()
	assert.Nil(t, err)
}

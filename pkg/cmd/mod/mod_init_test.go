package mod

import (
	"os"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/gitutil"
)

func TestValidateWithEmptyModuleName(t *testing.T) {
	o := &InitOptions{}
	err := o.Validate([]string{})
	assert.Error(t, err)
	assert.Equal(t, "module Name is empty", err.Error())
}

func TestValidateWithNonExistentPath(t *testing.T) {
	o := &InitOptions{}
	err := o.Validate([]string{"my-app", "/non/existent/Path"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create module directory")
}

func TestValidateWithExistingPath(t *testing.T) {
	o := &InitOptions{}
	err := o.Validate([]string{"my-app", os.TempDir()})
	assert.NoError(t, err)
}

func TestRunWithEmptyTemplateURL(t *testing.T) {
	o := &InitOptions{}
	o.Name = "my-app"
	o.Path = os.TempDir()
	err := o.Run()
	assert.NoError(t, err)
}

func TestRunWithInvalidTemplateURL(t *testing.T) {
	o := &InitOptions{}
	o.Name = "my-app"
	o.Path = os.TempDir()
	o.TemplateURL = "invalid-url"
	err := o.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to clone git repo")
}

func TestNewCmdInit(t *testing.T) {
	t.Run("validate error", func(t *testing.T) {
		// mock git clone
		mockey.Mock(gitutil.GitCloneOrPull).Return(nil).Build()
		cmd := NewCmdInit()
		cmd.SetArgs([]string{"fakeName"})
		err := cmd.Execute()
		assert.Nil(t, err)
	})
}

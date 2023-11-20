package build

import (
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdCompile(t *testing.T) {
	m1 := mockey.Mock((*Options).Complete).To(func(o *Options, args []string) error {
		o.Output = "stdout"
		return nil
	}).Build()
	m2 := mockey.Mock((*Options).Run).To(func(*Options) error {
		return nil
	}).Build()
	defer m1.UnPatch()
	defer m2.UnPatch()

	t.Run("compile success", func(t *testing.T) {
		cmd := NewCmdBuild()
		err := cmd.Execute()
		assert.Nil(t, err)
	})
}

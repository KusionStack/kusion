package compile

import (
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdCompile(t *testing.T) {
	m1 := mockey.Mock((*CompileOptions).Complete).To(func(o *CompileOptions, args []string) {
		o.Output = "stdout"
	}).Build()
	m2 := mockey.Mock((*CompileOptions).Run).To(func(*CompileOptions) error {
		return nil
	}).Build()
	defer m1.UnPatch()
	defer m2.UnPatch()

	t.Run("compile success", func(t *testing.T) {
		cmd := NewCmdCompile()
		err := cmd.Execute()
		assert.Nil(t, err)
	})
}

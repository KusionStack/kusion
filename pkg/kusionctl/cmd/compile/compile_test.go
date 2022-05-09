package compile

import (
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	kcl "kusionstack.io/kclvm-go"

	"kusionstack.io/kusion/pkg/compile"
)

func TestNewCmdCompile(t *testing.T) {
	t.Run("compile success", func(t *testing.T) {
		monkey.Patch(
			compile.Compile,
			func(workDir string,
				filenames, settings, arguments, overrides []string,
				disableNone bool,
				overrideAST bool) (*compile.CompileResult, error) {
				return &compile.CompileResult{
					Documents: []kcl.KCLResult{
						map[string]interface{}{
							"str":    "v1",
							"int":    2,
							"bool":   false,
							"struct": struct{}{},
						},
					},
				}, nil
			},
		)
		defer monkey.UnpatchAll()

		cmd := NewCmdCompile()
		err := cmd.Execute()
		assert.Nil(t, err)
	})
}

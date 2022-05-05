package check

import (
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"kusionstack.io/KCLVM/kclvm-go/api/kcl"

	"kusionstack.io/kusion/pkg/compile"
)

func TestNewCmdCheck(t *testing.T) {
	t.Run("", func(t *testing.T) {
		monkey.Patch(
			compile.Compile,
			func(workDir string,
				filenames, settings, arguments, overrides []string,
				disableNone bool,
				overrideAST bool,
			) (*compile.CompileResult, error) {
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
		cmd := NewCmdCheck()
		err := cmd.Execute()
		assert.Nil(t, err)
	})
}

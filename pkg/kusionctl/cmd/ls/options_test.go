package ls

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLsOptions_Run(t *testing.T) {
	o := &LsOptions{
		workDir: project.Path,
		Level:   2,
	}
	t.Run("json output", func(t *testing.T) {
		o.OutputFormat = "json"
		err := o.Run()
		assert.Nil(t, err)
	})

	t.Run("yaml output", func(t *testing.T) {
		o.OutputFormat = "yaml"
		err := o.Run()
		assert.Nil(t, err)
	})
}

package terminal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultUI(t *testing.T) {
	t.Run("init default ui", func(t *testing.T) {
		ui := DefaultUI()
		assert.NotNil(t, ui)
	})
}

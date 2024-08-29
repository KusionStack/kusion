package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestNewCmdRes(t *testing.T) {
	t.Run("successfully get resource help", func(t *testing.T) {
		streams, _, _, _ := genericiooptions.NewTestIOStreams()

		cmd := NewCmdRes(streams)
		assert.NotNil(t, cmd)
	})
}

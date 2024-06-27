package rel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestNewCmdRel(t *testing.T) {
	t.Run("successfully get release help", func(t *testing.T) {
		streams, _, _, _ := genericiooptions.NewTestIOStreams()

		cmd := NewCmdRel(streams)
		assert.NotNil(t, cmd)
	})
}

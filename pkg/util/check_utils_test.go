package util

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckArgument(t *testing.T) {
	t.Run("no panic", func(t *testing.T) {
		CheckArgument(true, "no panic")
	})
	t.Run("panic", func(t *testing.T) {
		defer func() {
			err := recover()
			assert.NotNil(t, err)
		}()
		CheckArgument(false, "panic mes")
	})
}

func TestCheckNotNil(t *testing.T) {
	t.Run("no panic", func(t *testing.T) {
		CheckNotNil(struct{}{}, "no panic")
	})
	t.Run("panic", func(t *testing.T) {
		defer func() {
			err := recover()
			assert.NotNil(t, err)
		}()
		CheckNotNil(nil, "panic")
	})
}

func TestCheckNotError(t *testing.T) {
	t.Run("no panic", func(t *testing.T) {
		CheckNotError(nil, "no panic")
	})
	t.Run("panic", func(t *testing.T) {
		defer func() {
			err := recover()
			assert.NotNil(t, err)
		}()
		CheckNotError(errors.New("panic"), "panic")
	})
}

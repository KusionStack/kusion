package util

import (
	"errors"
	"testing"
)

func TestRecoverErr(t *testing.T) {
	t.Run("recover error panic", func(t *testing.T) {
		err := errors.New("error")
		defer RecoverErr(&err)
		panic(err)
	})
	t.Run("recover string panic", func(t *testing.T) {
		var err error
		defer RecoverErr(&err)
		panic("error string")
	})
	t.Run("recover unknown panic", func(t *testing.T) {
		var err error
		defer RecoverErr(&err)
		panic(123)
	})
}

func TestCheckErr(t *testing.T) {
	t.Run("check error", func(t *testing.T) {
		defer func() {
			_ = recover()
		}()
		err := errors.New("error")
		CheckErr(err)
	})
}

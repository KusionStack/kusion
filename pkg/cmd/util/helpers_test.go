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

func TestParseClusterArgument(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "parse_cluster",
			args: args{args: []string{"cluster=fake"}},
			want: "fake",
		},
		{name: "parse_empty", args: args{args: nil}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseClusterArgument(tt.args.args); got != tt.want {
				t.Errorf("ParseClusterArgument() = %v, want %v", got, tt.want)
			}
		})
	}
}

package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_zapLogger(t *testing.T) {
	t.Run("log debug", func(t *testing.T) {
		Debug("foo")
	})
	t.Run("log debugf", func(t *testing.T) {
		Debugf("%s", "foo")
	})
	t.Run("log info", func(t *testing.T) {
		Info("foo")
	})
	t.Run("log infof", func(t *testing.T) {
		Infof("%s", "foo")
	})
	t.Run("log warn", func(t *testing.T) {
		Warn("foo")
	})
	t.Run("log warnf", func(t *testing.T) {
		Warnf("%s", "foo")
	})
	t.Run("log error", func(t *testing.T) {
		Error("foo")
	})
	t.Run("log errorf", func(t *testing.T) {
		Errorf("%s", "foo")
	})
	t.Run("log panic", func(t *testing.T) {
		defer func() {
			recover()
		}()
		Panic("foo")
	})
	t.Run("log panicf", func(t *testing.T) {
		defer func() {
			recover()
		}()
		Panicf("%s", "foo")
	})
	t.Run("log with", func(t *testing.T) {
		With("foo", "bar")
		With("count", 12)
	})
	t.Run("set level", func(t *testing.T) {
		SetLevel(FATAL)
		SetLevel(ERROR)
		SetLevel(WARN)
		SetLevel(INFO)
		SetLevel(DEBUG)
		SetLevel(8)
	})
	t.Run("get log dir config", func(t *testing.T) {
		dir := GetLogDir()
		assert.NotNil(t, dir)
	})
	t.Run("get logger", func(t *testing.T) {
		logger := GetLogger()
		assert.NotNil(t, logger)
	})
}

func TestGetLevelFromStr(t *testing.T) {
	type args struct {
		level string
	}
	tests := []struct {
		name string
		args args
		want Level
	}{
		{
			name: "DEBUG",
			args: args{
				level: "DEBUG",
			},
			want: DEBUG,
		},
		{
			name: "INFO",
			args: args{
				level: "INFO",
			},
			want: INFO,
		},
		{
			name: "WARN",
			args: args{
				level: "WARN",
			},
			want: WARN,
		},
		{
			name: "ERROR",
			args: args{
				level: "ERROR",
			},
			want: ERROR,
		},
		{
			name: "FATAL",
			args: args{
				level: "FATAL",
			},
			want: FATAL,
		},
		{
			name: "PANIC",
			args: args{
				level: "PANIC",
			},
			want: INFO,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLevelFromStr(tt.args.level); got != tt.want {
				t.Errorf("GetLevelFromStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

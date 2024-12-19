package util

import (
	"strings"

	"github.com/go-chi/httplog/v2"
)

func LogToAll(sysLogger *httplog.Logger, runLogger *httplog.Logger, level string, message string, args ...any) {
	switch strings.ToLower(level) {
	case "info":
		sysLogger.Info(message, args...)
		runLogger.Info(message, args...)
	case "error":
		sysLogger.Error(message, args...)
		runLogger.Error(message, args...)
	case "warn":
		sysLogger.Warn(message, args...)
		runLogger.Warn(message, args...)
	case "debug":
		sysLogger.Debug(message, args...)
		runLogger.Debug(message, args...)
	default:
		sysLogger.Error("unknown log level", "level", level)
	}
}

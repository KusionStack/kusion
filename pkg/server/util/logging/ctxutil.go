package util

import (
	"bytes"
	"context"

	"github.com/go-chi/httplog/v2"
	"kusionstack.io/kusion/pkg/server/middleware"
)

// GetLogger returns the logger from the given context.
//
// Example:
//
//	logger := ctxlogutil.GetLogger(ctx)
func GetLogger(ctx context.Context) *httplog.Logger {
	if logger, ok := ctx.Value(middleware.APILoggerKey).(*httplog.Logger); ok {
		return logger
	}

	return httplog.NewLogger("DefaultLogger")
}

// GetRunLogger returns the run logger from the given context.
func GetRunLogger(ctx context.Context) *httplog.Logger {
	if logger, ok := ctx.Value(middleware.RunLoggerKey).(*httplog.Logger); ok {
		return logger
	}

	return httplog.NewLogger("DefaultRunLogger")
}

// GetRunLoggerBuffer returns the run logger buffer from the given context.
func GetRunLoggerBuffer(ctx context.Context) *bytes.Buffer {
	if buffer, ok := ctx.Value(middleware.RunLoggerBufferKey).(*bytes.Buffer); ok {
		return buffer
	}

	return &bytes.Buffer{}
}

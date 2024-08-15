package util

import (
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

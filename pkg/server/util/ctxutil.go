package util

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
	"kusionstack.io/kusion/pkg/server/middleware"
)

// GetLogger returns the logger from the given context.
//
// Example:
//
//	logger := ctxutil.GetLogger(ctx)
func GetLogger(ctx context.Context) logr.Logger {
	if logger, ok := ctx.Value(middleware.APILoggerKey).(logr.Logger); ok {
		return logger
	}

	return klog.NewKlogr()
}

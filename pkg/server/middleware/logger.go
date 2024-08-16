package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/httplog/v2"
	"k8s.io/klog/v2"
)

// APILoggerKey is a context key used for associating a logger with a request.
var APILoggerKey = &contextKey{"logger"}

func InitLogger(logFilePath string, name string) *httplog.Logger {
	logWriter, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		klog.Fatalf("Failed to open log file: %v", err)
	}
	logger := httplog.NewLogger(name, httplog.Options{
		LogLevel:        slog.LevelInfo,
		Concise:         true,
		TimeFieldFormat: time.RFC3339,
		Writer:          logWriter,
		RequestHeaders:  true,
		Trace: &httplog.TraceOptions{
			HeaderTrace: "x-kusion-trace",
		},
	})
	return logger
}

// APILoggerMiddleware injects a logger, configured with a request ID,
// into the request context for use throughout the request's lifecycle.
func APILoggerMiddleware(logFile string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// Retrieve the request ID from the context and create a logger with it.
			if requestID := GetTraceID(ctx); len(requestID) > 0 {
				// Set the output file for klog
				logger := InitLogger(logFile, requestID)
				ctx = context.WithValue(ctx, APILoggerKey, logger)
			}
			// Continue serving the request with the new context.
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// DefaultLogger is a middleware that provides basic request logging using chi's
// built-in Logger middleware.
func DefaultLoggerMiddleware(logFile string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		logger := InitLogger(logFile, "DefaultLogger")
		return httplog.RequestLogger(logger)(next)
	}
}

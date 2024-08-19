package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTraceID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := GetTraceID(r.Context())
		w.Write([]byte(traceID))
	})

	// Create a new request with a mock response recorder
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Create a context and add the trace ID header
	ctx := context.Background()
	req = req.WithContext(ctx)
	req.Header.Set("x-kusion-trace", "test-trace-id")

	// Call the middleware with the handler
	TraceID(handler).ServeHTTP(w, req)

	// Assert that the response body contains the trace ID
	assert.Equal(t, "test-trace-id", w.Body.String())
}

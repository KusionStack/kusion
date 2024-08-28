package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokenAuthMiddleware(t *testing.T) {
	// Define the key map and whitelist
	keyMap := map[string]interface{}{
		"key1": "publicKey1",
		"key2": "publicKey2",
	}
	whitelist := []string{"subject1", "subject2"}

	// Create a test request with a valid token
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer inValidToken")

	// Create a test response recorder
	rr := httptest.NewRecorder()

	// Create a mock handler
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create the middleware handler
	middlewareHandler := TokenAuthMiddleware(keyMap, whitelist, "test.log")(mockHandler)

	// Serve the request through the middleware
	middlewareHandler.ServeHTTP(rr, req)

	// Assert that the response status code is 200 OK
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

package middleware

import (
	"net/http"
)

// ReadOnlyMode disallows non-GET requests in read-only mode.
func ReadOnlyMode(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "The server is currently in read-only mode.", http.StatusMethodNotAllowed)
			return
		}

		// If the request method is allowed, pass the request to the next handler.
		next.ServeHTTP(w, r)
	})
}

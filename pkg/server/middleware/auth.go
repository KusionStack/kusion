package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/jwtauth/v5"
	jwt "github.com/golang-jwt/jwt/v5"
)

// This is the middleware function that verifies the JWT against a JWKS KeyMap
func TokenAuthMiddleware(keyMap map[string]any, whitelist []string, logFilePath string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use the proper key to verify the token
			logger := GetMiddlewareLogger(r.Context(), logFilePath)
			tokenString := jwtauth.TokenFromHeader(r)
			t, err := jwt.Parse(tokenString, func(tok *jwt.Token) (interface{}, error) {
				// Get the kid from the token header
				kid, ok := tok.Header["kid"].(string)
				if !ok {
					return nil, fmt.Errorf("%w: could not find and convert kid in JWT header to string", errors.New("the JWT has an invalid kid"))
				}
				logger.Info("kid found", "kid", kid)
				publicKey, ok := keyMap[kid]
				if !ok {
					return nil, fmt.Errorf("%w: could not find the public key to verify signature", errors.New("the kid cannot be found in the JWKS"))
				}
				return publicKey, nil
			})
			if err != nil {
				logger.Info("error parsing token", "error", err)
				logger.Info("request is denied")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			// Get subject from claims
			subject, err := t.Claims.GetSubject()
			if err != nil {
				logger.Info("error getting subject", "error", err)
				logger.Info("request is denied")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			logger.Info("subject found", "subject", subject)

			// Get profile from claims
			profile, ok := t.Claims.(jwt.MapClaims)["profile"]
			if ok {
				logger.Info("profile found", "profile", profile)
			}

			// Check if the org name is in the whitelist
			orgName := profile.(map[string]any)["org_name"]
			orgType := profile.(map[string]any)["org_type"]

			// Check if the subject is in the whitelist
			if !SubjectWhitelisted(subject, orgType.(string), orgName.(string), whitelist) {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			// Token is authenticated, pass it through
			logger.Info("request is authorized")
			next.ServeHTTP(w, r)
		})
	}
}

func SubjectWhitelisted(subject, orgType, orgName string, whitelist []string) bool {
	// iterate through whitelist and return true of any matches with subject
	for _, whitelistedItem := range whitelist {
		if subject == whitelistedItem || (orgType == "application" && orgName == whitelistedItem) {
			return true
		}
	}
	return false
}

func GetMiddlewareLogger(ctx context.Context, logFile string) *httplog.Logger {
	if logger, ok := ctx.Value(APILoggerKey).(*httplog.Logger); ok {
		return logger
	}
	logger := InitLogger(logFile, "DefaultLogger")
	return logger
}

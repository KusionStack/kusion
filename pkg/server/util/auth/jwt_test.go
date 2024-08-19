package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func TestGetSubjectFromUnverifiedJWTToken(t *testing.T) {
	ctx := context.Background()

	signingKey := []byte("my-secret-key") // Replace with your secret key
	expectedSubject := "1234567890"

	// Create the token
	claims := &jwt.RegisteredClaims{
		Subject:   expectedSubject,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)), // Token valid for 72 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate the encoded token
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		t.Errorf("error generating token: %v", err)
	}

	// Create a mock request with the IAM_TOKEN cookie
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	cookie := &http.Cookie{
		Name:  "IAM_TOKEN",
		Value: tokenString,
	}
	req.AddCookie(cookie)

	subject, err := GetSubjectFromUnverifiedJWTToken(ctx, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if subject != expectedSubject {
		t.Errorf("expected subject %q, got %q", expectedSubject, subject)
	}
}

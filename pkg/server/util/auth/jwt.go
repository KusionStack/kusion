package auth

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/jwk"
	httputil "kusionstack.io/kusion/pkg/server/util/http"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

func GetJWKSMapFromIAM(ctx context.Context, keyType string) (map[string]any, error) {
	logger := logutil.GetLogger(ctx)
	keyMap := make(map[string]any)
	baseURL := os.Getenv("IAM_URL")
	// TODO: This does not change frequently, we should not need to pull it in every request
	verifyURL := fmt.Sprintf("%s/api/auth/oidc/.well-known/jwks.json", baseURL)

	// Default to RSA key type
	if keyType == "" {
		keyType = "RSA"
	}

	// Get JWKS from IAM
	req, err := http.NewRequest("GET", verifyURL, nil)
	if err != nil {
		logger.Info("Error creating request:", "error", err)
		return nil, err
	}

	resp, body, err := httputil.ProcessResponse(ctx, req)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get jwk from IAM: %s", string(body))
	}

	// Parse the key set
	set, err := jwk.Parse(body)
	if err != nil {
		return nil, err
	}

	// Populate keyMap
	for it := set.Iterate(context.Background()); it.Next(context.Background()); {
		pair := it.Pair()
		key := pair.Value.(jwk.Key)
		kid := pair.Value.(jwk.Key).KeyID()

		var rawkey interface{} // This is the raw key, like *rsa.PrivateKey or *ecdsa.PrivateKey
		if err := key.Raw(&rawkey); err != nil {
			log.Printf("failed to create public key: %s", err)
			return nil, err
		}

		// We know this is an RSA Key
		if strings.ToUpper(keyType) == "RSA" {
			rsa, ok := rawkey.(*rsa.PublicKey)
			if !ok {
				panic(fmt.Sprintf("expected rsa key, got %T", rawkey))
			}
			keyMap[kid] = rsa
		}
	}
	return keyMap, nil
}

// GetSubjectFromUnverifiedJWTToken returns the subject from an unverified JWT token.
// This is a temp solution until I know where to get the public key to verify the token.
// This is used for parsing the session information in cookies from frontend only.
func GetSubjectFromUnverifiedJWTToken(ctx context.Context, r *http.Request) (string, error) {
	logger := logutil.GetLogger(ctx)

	iamToken, err := r.Cookie("IAM_TOKEN")
	if err == http.ErrNoCookie {
		return "", nil
	} else if err != nil {
		return "", err
	}

	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(iamToken.Value, jwt.MapClaims{})
	if err != nil {
		return "", err
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		return "", err
	}
	logger.Info("token subject:", "subject", subject)

	return subject, nil
}

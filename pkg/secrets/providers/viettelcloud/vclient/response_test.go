package vclient

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
)

func TestParseSecretManagerSecretsListResponse(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		statusCode     int
		contentType    string
		expectedResult *SecretManagerSecretsListResponse
		expectError    bool
	}{
		{
			name:         "Valid 200 Response",
			responseBody: "{\"results\": []}",
			statusCode:   http.StatusOK,
			contentType:  "application/json",
			expectedResult: &SecretManagerSecretsListResponse{
				JSON200: &PaginatedSecretListList{
					Count:    nil,
					Next:     nil,
					Previous: nil,
					Results:  &[]SecretList{},
				},
			},
			expectError: false,
		},
		{
			name:         "Valid 400 Response",
			responseBody: "{\"error\": \"Bad Request\"}",
			statusCode:   http.StatusBadRequest,
			contentType:  "application/json",
			expectedResult: &SecretManagerSecretsListResponse{
				JSON400: &ErrorResponse{
					Union: json.RawMessage(nil),
				},
			},
			expectError: false,
		},
		{
			name:         "Invalid JSON Response",
			responseBody: `invalid json`,
			statusCode:   http.StatusOK,
			contentType:  "application/json",
			expectedResult: &SecretManagerSecretsListResponse{
				JSON200: nil,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header: map[string][]string{
					"Content-Type": {tt.contentType},
				},
				Body: io.NopCloser(bytes.NewBufferString(tt.responseBody)),
			}
			parsedResponse, err := ParseSecretManagerSecretsListResponse(resp)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.JSON200, parsedResponse.JSON200)
				assert.Equal(t, tt.expectedResult.JSON400, parsedResponse.JSON400)
			}
		})
	}
}

func TestParseSecretManagerSecretsRetrieveResponse(t *testing.T) {
	id, _ := uuid.Parse("123e4567-e89b-12d3-a456-426614174000")
	tests := []struct {
		name           string
		responseBody   string
		statusCode     int
		contentType    string
		expectedResult *SecretManagerSecretsRetrieveResponse
		expectError    bool
	}{
		{
			name:         "Valid 200 Response",
			responseBody: "{\"id\": \"123e4567-e89b-12d3-a456-426614174000\"}",
			statusCode:   http.StatusOK,
			contentType:  "application/json",
			expectedResult: &SecretManagerSecretsRetrieveResponse{
				JSON200: &SecretRetrieve{
					ID: &id,
				},
			},
			expectError: false,
		},
		{
			name:         "Valid 400 Response",
			responseBody: `{"error": "Bad Request"}`,
			statusCode:   http.StatusBadRequest,
			contentType:  "application/json",
			expectedResult: &SecretManagerSecretsRetrieveResponse{
				JSON400: &ErrorResponse{
					Union: json.RawMessage(nil),
				},
			},
			expectError: false,
		},
		{
			name:         "Invalid JSON Response",
			responseBody: `invalid json`,
			statusCode:   http.StatusOK,
			contentType:  "application/json",
			expectedResult: &SecretManagerSecretsRetrieveResponse{
				JSON200: nil,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header: map[string][]string{
					"Content-Type": {tt.contentType},
				},
				Body: io.NopCloser(bytes.NewBufferString(tt.responseBody)),
			}
			parsedResponse, err := ParseSecretManagerSecretsRetrieveResponse(resp)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.JSON200, parsedResponse.JSON200)
				assert.Equal(t, tt.expectedResult.JSON400, parsedResponse.JSON400)
			}
		})
	}
}

package vclient

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Mock HTTP Client
type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestSecretManagerSecretsList(t *testing.T) {
	mockResponse := `{
		"count": 4,
		"next": null,
		"previous": null,
		"results": [
			{
				"id": "973fbcad-0571-4145-9fea-bafd574580ef",
				"name": "secret-00",
				"created_at": "2024-08-13T03:36:13.175203Z"
			},
			{
				"id": "7ef5e845-9f0a-40b5-81e1-5fdb57da1f2f",
				"name": "secret-01",
				"created_at": "2024-08-13T02:41:34.721359Z"
			},
			{
				"id": "c9b96dea-1ee5-4f01-8343-bb288df0c6dc",
				"name": "secret-02",
				"created_at": "2024-08-05T18:27:04.169141Z"
			},
			{
				"id": "6d4d291e-d6dc-441a-9aad-9ef353ebdd79",
				"name": "secret-03",
				"created_at": "2024-07-11T07:18:38.146912Z"
			}
		]
	}`

	testCase := map[string]struct {
		mockClient *MockClient
		params     *SecretManagerSecretsListParams
		expected   *http.Response
		expectErr  error
	}{
		"SecretsList_Success": {
			mockClient: &MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 200,
						Header: map[string][]string{
							"Content-Type": {"application/json"},
						},
						Body: io.NopCloser(bytes.NewBufferString(mockResponse)),
					}, nil
				},
			},
			params: &SecretManagerSecretsListParams{},
			expected: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body: io.NopCloser(bytes.NewBufferString(mockResponse)),
			},
			expectErr: nil,
		},
		"SecretsList_Failure": {
			mockClient: &MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 500,
						Header: map[string][]string{
							"Content-Type": {"application/json"},
						},
						Body: io.NopCloser(bytes.NewBufferString(`{"error": "internal server error"}`)),
					}, errors.New("internal server error")
				},
			},
			params: &SecretManagerSecretsListParams{},
			expected: &http.Response{
				StatusCode: 500,
				Header: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body: io.NopCloser(bytes.NewBufferString(`{"error": "internal server error"}`)),
			},
			expectErr: errors.New("internal server error"),
		},
		"SecretsList_EmptyResponse": {
			mockClient: &MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 200,
						Header: map[string][]string{
							"Content-Type": {"application/json"},
						},
						Body: io.NopCloser(bytes.NewBufferString(`{}`)),
					}, nil
				},
			},
			params: &SecretManagerSecretsListParams{},
			expected: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body: io.NopCloser(bytes.NewBufferString(`{}`)),
			},
			expectErr: nil,
		},
	}

	for name, tc := range testCase {
		t.Run(name, func(t *testing.T) {
			client := &Client{
				Server: "http://example.com",
				Client: tc.mockClient,
			}
			resp, err := client.SecretManagerSecretsList(context.Background(), tc.params)
			defer func() { _ = resp.Body.Close() }()
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expected, resp)
		})
	}
}

func TestSecretManagerSecretsRetrieve(t *testing.T) {
	mockResponse := `{
		"id": "7ef5e845-9f0a-40b5-81e1-5fdb57da1f2f",
		"name": "secret-01",
		"created_at": "2024-08-13T02:41:34.721359Z",
		"secret": {
			"host": "mysql.example.com",
			"password": "password",
			"username": "username"
		},
		"metadata": null
	}`
	testCase := map[string]struct {
		mockClient *MockClient
		id         uuid.UUID
		params     *SecretManagerSecretsRetrieveParams
		expected   *http.Response
		expectErr  error
	}{
		"SecretsRetrieve_Success": {
			mockClient: &MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 200,
						Header: map[string][]string{
							"Content-Type": {"application/json"},
						},
						Body: io.NopCloser(bytes.NewBufferString(mockResponse)),
					}, nil
				},
			},
			id:     uuid.New(),
			params: &SecretManagerSecretsRetrieveParams{},
			expected: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body: io.NopCloser(bytes.NewBufferString(mockResponse)),
			},
			expectErr: nil,
		},
		"SecretsRetrieve_Failure": {
			mockClient: &MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 500,
						Header: map[string][]string{
							"Content-Type": {"application/json"},
						},
						Body: io.NopCloser(bytes.NewBufferString(`{"error": "internal server error"}`)),
					}, errors.New("internal server error")
				},
			},
			id:     uuid.New(),
			params: &SecretManagerSecretsRetrieveParams{},
			expected: &http.Response{
				StatusCode: 500,
				Header: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body: io.NopCloser(bytes.NewBufferString(`{"error": "internal server error"}`)),
			},
			expectErr: errors.New("internal server error"),
		},
		"SecretsRetrieve_Failure_404": {
			mockClient: &MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 404,
						Header: map[string][]string{
							"Content-Type": {"application/json"},
						},
						Body: io.NopCloser(bytes.NewBufferString(`{"error": "not found"}`)),
					}, errors.New("not found")
				},
			},
			id:     uuid.New(),
			params: &SecretManagerSecretsRetrieveParams{},
			expected: &http.Response{
				StatusCode: 404,
				Header: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body: io.NopCloser(bytes.NewBufferString(`{"error": "not found"}`)),
			},
			expectErr: errors.New("not found"),
		},
	}

	for name, tc := range testCase {
		t.Run(name, func(t *testing.T) {
			client := &Client{
				Server: "http://example.com",
				Client: tc.mockClient,
			}
			resp, err := client.SecretManagerSecretsRetrieve(context.Background(), tc.id, tc.params)
			defer func() { _ = resp.Body.Close() }()
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expected, resp)
		})
	}
}

func TestSecretManagerSecretsListWithResponse(t *testing.T) {
	mockResponse := "{\"count\":4,\"next\":null,\"previous\":null,\"results\":[{\"id\":\"973fbcad-0571-4145-9fea-bafd574580ef\",\"name\":\"ide-platform\",\"created_at\":\"2024-08-13T03:36:13.175203Z\"},{\"id\":\"7ef5e845-9f0a-40b5-81e1-5fdb57da1f2f\",\"name\":\"mysql-secrets\",\"created_at\":\"2024-08-13T02:41:34.721359Z\"},{\"id\":\"c9b96dea-1ee5-4f01-8343-bb288df0c6dc\",\"name\":\"test_terraform_secret\",\"created_at\":\"2024-08-05T18:27:04.169141Z\"},{\"id\":\"6d4d291e-d6dc-441a-9aad-9ef353ebdd79\",\"name\":\"terraform_secret01\",\"created_at\":\"2024-07-11T07:18:38.146912Z\"}]}"
	testCase := map[string]struct {
		mockClient *MockClient
		params     *SecretManagerSecretsListParams
		expected   *SecretManagerSecretsListResponse
		expectErr  error
	}{
		"SecretsList_Success": {
			mockClient: &MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 200,
						Header: map[string][]string{
							"Content-Type": {"application/json"},
						},
						Body: io.NopCloser(bytes.NewBufferString(mockResponse)),
					}, nil
				},
			},
			params: &SecretManagerSecretsListParams{},
			expected: parseSecretManagerSecretsListResponse(&http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body: io.NopCloser(bytes.NewBufferString(mockResponse)),
			}),
			expectErr: nil,
		},
		"SecretsList_Failure": {
			mockClient: &MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 500,
						Header: map[string][]string{
							"Content-Type": {"application/json"},
						},
						Body: io.NopCloser(bytes.NewBufferString(`{"error": "internal server error"}`)),
					}, errors.New("internal server error")
				},
			},
			params: &SecretManagerSecretsListParams{},
			expected: parseSecretManagerSecretsListResponse(&http.Response{
				StatusCode: 500,
				Header: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body: io.NopCloser(bytes.NewBufferString(`{"error": "internal server error"}`)),
			}),
			expectErr: errors.New("internal server error"),
		},
		"SecretsList_EmptyResponse": {
			mockClient: &MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 200,
						Header: map[string][]string{
							"Content-Type": {"application/json"},
						},
						Body: io.NopCloser(bytes.NewBufferString(`{}`)),
					}, nil
				},
			},
			params: &SecretManagerSecretsListParams{},
			expected: parseSecretManagerSecretsListResponse(&http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body: io.NopCloser(bytes.NewBufferString(`{}`)),
			}),
			expectErr: nil,
		},
	}

	for name, tc := range testCase {
		t.Run(name, func(t *testing.T) {
			client := &ClientWithResponses{
				ClientInterface: &Client{
					Server: "http://example.com",
					Client: tc.mockClient,
				},
			}
			params := &SecretManagerSecretsListParams{ProjectID: uuid.New()}
			resp, err := client.SecretManagerSecretsListWithResponse(context.Background(), params)
			assert.Equal(t, tc.expectErr, err)
			if err == nil {
				assert.NotNil(t, resp.JSON200)
				assert.Equal(t, tc.expected, resp)
			}
		})
	}
}

func TestSecretManagerSecretsRetrieveWithResponse(t *testing.T) {
	mockListResponse := "{\"count\":4,\"next\":null,\"previous\":null,\"results\":[{\"id\":\"973fbcad-0571-4145-9fea-bafd574580ef\",\"name\":\"secret-00\",\"created_at\":\"2024-08-13T03:36:13.175203Z\"},{\"id\":\"7ef5e845-9f0a-40b5-81e1-5fdb57da1f2f\",\"name\":\"secret-01\",\"created_at\":\"2024-08-13T02:41:34.721359Z\"},{\"id\":\"c9b96dea-1ee5-4f01-8343-bb288df0c6dc\",\"name\":\"secret-02\",\"created_at\":\"2024-08-05T18:27:04.169141Z\"},{\"id\":\"6d4d291e-d6dc-441a-9aad-9ef353ebdd79\",\"name\":\"secret-03\",\"created_at\":\"2024-07-11T07:18:38.146912Z\"}]}"
	mockRetrieveResponse := "{\"id\":\"7ef5e845-9f0a-40b5-81e1-5fdb57da1f2f\",\"name\":\"secret-01\",\"created_at\":\"2024-08-13T02:41:34.721359Z\",\"secret\":{\"host\":\"mysql.example.com\",\"password\":\"password\",\"username\":\"username\"},\"metadata\":null}"
	testCase := map[string]struct {
		mockClient *MockClient
		name       string
		params     *SecretManagerSecretsRetrieveParams
		expected   *SecretManagerSecretsRetrieveResponse
		expectErr  error
	}{
		"SecretsRetrieve_Success": {
			mockClient: &MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					if req.URL.Path == prefix {
						return &http.Response{
							StatusCode: 200,
							Header: map[string][]string{
								"Content-Type": {"application/json"},
							},
							Body: io.NopCloser(bytes.NewBufferString(mockListResponse)),
						}, nil
					}
					return &http.Response{
						StatusCode: 200,
						Header: map[string][]string{
							"Content-Type": {"application/json"},
						},
						Body: io.NopCloser(bytes.NewBufferString(mockRetrieveResponse)),
					}, nil
				},
			},
			name:   "secret-01",
			params: &SecretManagerSecretsRetrieveParams{},
			expected: parseSecretManagerSecretsRetrieveResponse(&http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body: io.NopCloser(bytes.NewBufferString(mockRetrieveResponse)),
			}),
			expectErr: nil,
		},
		"SecretsRetrieve_Failure": {
			mockClient: &MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					if req.URL.Path == prefix {
						return &http.Response{
							StatusCode: 200,
							Header: map[string][]string{
								"Content-Type": {"application/json"},
							},
							Body: io.NopCloser(bytes.NewBufferString(mockListResponse)),
						}, nil
					}
					return &http.Response{
						StatusCode: 500,
						Header: map[string][]string{
							"Content-Type": {"application/json"},
						},
						Body: io.NopCloser(bytes.NewBufferString(`{"error": "internal server error"}`)),
					}, errors.New("internal server error")
				},
			},
			name:   "secret-01",
			params: &SecretManagerSecretsRetrieveParams{},
			expected: parseSecretManagerSecretsRetrieveResponse(&http.Response{
				StatusCode: 500,
				Header: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body: io.NopCloser(bytes.NewBufferString(`{"error": "internal server error"}`)),
			}),
			expectErr: errors.New("internal server error"),
		},
	}

	for name, tc := range testCase {
		t.Run(name, func(t *testing.T) {
			client := &ClientWithResponses{
				ClientInterface: &Client{
					Server: "http://example.com",
					Client: tc.mockClient,
				},
			}
			params := &SecretManagerSecretsRetrieveParams{ProjectID: uuid.New()}
			resp, err := client.SecretManagerSecretsRetrieveWithResponse(context.Background(), tc.name, params)
			assert.Equal(t, tc.expectErr, err)
			if err == nil {
				assert.NotNil(t, resp.JSON200)
				assert.Equal(t, tc.expected, resp)
			}
		})
	}
}

func parseSecretManagerSecretsListResponse(rsp *http.Response) *SecretManagerSecretsListResponse {
	rs, err := ParseSecretManagerSecretsListResponse(rsp)
	if err != nil {
		return nil
	}
	return rs
}

func parseSecretManagerSecretsRetrieveResponse(rsp *http.Response) *SecretManagerSecretsRetrieveResponse {
	rs, err := ParseSecretManagerSecretsRetrieveResponse(rsp)
	if err != nil {
		return nil
	}
	return rs
}

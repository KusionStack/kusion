package vclient

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type SecurityProviderPATToken struct {
	token string
}

func NewSecurityProviderPATToken(token string) (*SecurityProviderPATToken, error) {
	return &SecurityProviderPATToken{
		token: token,
	}, nil
}

func (s *SecurityProviderPATToken) Intercept(ctx context.Context, req *http.Request) error {
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", s.token))
	return nil
}

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// HTTPRequestDoer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HTTPRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The base URL for the service
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HTTPRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) SecretManagerSecretsList(ctx context.Context, params *SecretManagerSecretsListParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewSecretManagerSecretsListRequest(c.Server, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) SecretManagerSecretsRetrieve(ctx context.Context, id uuid.UUID, params *SecretManagerSecretsRetrieveParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewSecretManagerSecretsRetrieveRequest(c.Server, id, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// NewClient creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a vclient with sane default values
	client := Client{
		Server: server,
	}
	// mutate vclient and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// ClientInterface is the interface for the vclient
type ClientInterface interface {
	// SecretManagerSecretsList request
	SecretManagerSecretsList(ctx context.Context, params *SecretManagerSecretsListParams, reqEditors ...RequestEditorFn) (*http.Response, error)
	// SecretManagerSecretsRetrieve request
	SecretManagerSecretsRetrieve(ctx context.Context, id uuid.UUID, params *SecretManagerSecretsRetrieveParams, reqEditors ...RequestEditorFn) (*http.Response, error)
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// SecretManagerSecretsListWithResponse request returning *SecretManagerSecretsListResponse
func (c *ClientWithResponses) SecretManagerSecretsListWithResponse(ctx context.Context, params *SecretManagerSecretsListParams, reqEditors ...RequestEditorFn) (*SecretManagerSecretsListResponse, error) {
	rsp, err := c.SecretManagerSecretsList(ctx, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rsp.Body.Close() }()
	return ParseSecretManagerSecretsListResponse(rsp)
}

func (c *ClientWithResponses) SecretManagerSecretsRetrieveWithResponse(ctx context.Context, name string, params *SecretManagerSecretsRetrieveParams, reqEditors ...RequestEditorFn) (*SecretManagerSecretsRetrieveResponse, error) {
	listParams := &SecretManagerSecretsListParams{
		Name:      &name,
		ProjectID: params.ProjectID,
	}
	smsListResponse, err := c.SecretManagerSecretsListWithResponse(ctx, listParams, reqEditors...)
	if err != nil {
		return nil, err
	}
	if smsListResponse.JSON200 == nil {
		return nil, fmt.Errorf("response body is nil")
	}
	if smsListResponse.JSON200.Results == nil {
		return nil, fmt.Errorf("results is nil")
	}
	if smsListResponse.JSON400 != nil {
		return nil, fmt.Errorf("error response: %s", smsListResponse.JSON400.Union)
	}
	if smsListResponse.JSON401 != nil {
		return nil, fmt.Errorf("error response: %s", smsListResponse.JSON401.Union)
	}
	if smsListResponse.JSON403 != nil {
		return nil, fmt.Errorf("error response: %s", smsListResponse.JSON403.Union)
	}
	id := uuid.UUID{}
	for _, secret := range *smsListResponse.JSON200.Results {
		if secret.Name == name {
			id = *secret.ID
			break
		}
	}
	rsp, err := c.SecretManagerSecretsRetrieve(ctx, id, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rsp.Body.Close() }()
	return ParseSecretManagerSecretsRetrieveResponse(rsp)
}

package vclient

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type SecretManagerSecretsListResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *PaginatedSecretListList
	JSON400      *ErrorResponse
	JSON401      *ErrorResponse
	JSON403      *ErrorResponse
}

// Status returns HTTPResponse.Status
func (r SecretManagerSecretsListResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r SecretManagerSecretsListResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// ParseSecretManagerSecretsListResponse parses an HTTP response from a SecretManagerSecretsListWithResponse call
func ParseSecretManagerSecretsListResponse(rsp *http.Response) (*SecretManagerSecretsListResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &SecretManagerSecretsListResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest PaginatedSecretListList
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 401:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON401 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 403:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON403 = &dest
	}
	return response, nil
}

type SecretManagerSecretsRetrieveResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *SecretRetrieve
	JSON400      *ErrorResponse
	JSON401      *ErrorResponse
	JSON403      *ErrorResponse
	JSON404      *ErrorResponse
}

// Status returns HTTPResponse.Status
func (r SecretManagerSecretsRetrieveResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r SecretManagerSecretsRetrieveResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// ParseSecretManagerSecretsRetrieveResponse parses an HTTP response from a SecretManagerSecretsRetrieveWithResponse call
func ParseSecretManagerSecretsRetrieveResponse(rsp *http.Response) (*SecretManagerSecretsRetrieveResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &SecretManagerSecretsRetrieveResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest SecretRetrieve
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 401:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON401 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 403:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON403 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 404:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON404 = &dest
	}
	return response, nil
}

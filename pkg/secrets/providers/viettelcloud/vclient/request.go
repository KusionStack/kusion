package vclient

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/oapi-codegen/runtime"

	"github.com/google/uuid"
)

const prefix = "/secret-manager/secrets/"

// NewSecretManagerSecretsListRequest generates requests for SecretManagerSecretsList
func NewSecretManagerSecretsListRequest(server string, params *SecretManagerSecretsListParams) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := prefix
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	queryValues := queryURL.Query()

	if params.Name != nil {
		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "name", runtime.ParamLocationQuery, *params.Name); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}
	}

	if params.Page != nil {
		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "page", runtime.ParamLocationQuery, *params.Page); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}
	}

	if params.PageSize != nil {
		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "page_size", runtime.ParamLocationQuery, *params.PageSize); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}
	}

	queryURL.RawQuery = queryValues.Encode()

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	var projectID string

	projectID, err = runtime.StyleParamWithLocation("simple", false, "project-id", runtime.ParamLocationHeader, params.ProjectID)
	if err != nil {
		return nil, err
	}

	req.Header.Set("project-id", projectID)

	return req, nil
}

// NewSecretManagerSecretsRetrieveRequest generates requests for SecretManagerSecretsRetrieve
func NewSecretManagerSecretsRetrieveRequest(server string, id uuid.UUID, params *SecretManagerSecretsRetrieveParams) (*http.Request, error) {
	var err error

	var pathParam string

	pathParam, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("%s%s/", prefix, pathParam)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	var projectID string

	projectID, err = runtime.StyleParamWithLocation("simple", false, "project-id", runtime.ParamLocationHeader, params.ProjectID)
	if err != nil {
		return nil, err
	}

	req.Header.Set("project-id", projectID)

	return req, nil
}

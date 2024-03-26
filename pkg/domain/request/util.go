package request

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
)

// decode detects the correct decoder for use on an HTTP request and
// marshals into a given interface.
func decode(r *http.Request, payload interface{}) error {
	// Check if the content type is plain text, read it as such.
	contentType := render.GetRequestContentType(r)
	switch contentType {
	case render.ContentTypeJSON:
		// For non-plain text, decode the JSON body into the payload.
		if err := render.DecodeJSON(r.Body, payload); err != nil {
			return err
		}
	default:
		return errors.New("unsupported media type")
	}

	return nil
}

func (payload *CreateProjectRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *UpdateProjectRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *CreateStackRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *UpdateStackRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *CreateSourceRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *UpdateSourceRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *CreateOrganizationRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *UpdateOrganizationRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *CreateWorkspaceRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *UpdateWorkspaceRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *CreateBackendRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *UpdateBackendRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

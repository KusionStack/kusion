package request

import (
	"errors"
	"net/http"
	"regexp"

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

func validPath(path string) bool {
	// Validate project and stack path contains one or more capturing group
	// that contains a backslash with alphanumeric and underscore characters
	return !regexp.MustCompile(`^([\/a-zA-Z0-9_])+$`).MatchString(path)
}

func validName(name string) bool {
	// Validate project, stack and appconfig name contains only alphanumeric
	// and underscore characters
	return !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(name)
}

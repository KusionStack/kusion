package request

import "net/http"

// CreateSourceRequest represents the create request structure for
// source.
type CreateSourceRequest struct {
	// Name is the name of the source.
	Name string `json:"name" binding:"required"`
	// SourceProvider is the type of the source provider.
	SourceProvider string `json:"sourceProvider" binding:"required"`
	// Remote is the source URL, including scheme.
	Remote string `json:"remote" binding:"required"`
	// Description is a human-readable description of the source.
	Description string `json:"description"`
	// Labels are custom labels associated with the source.
	Labels []string `json:"labels"`
	// Owners is a list of owners for the source.
	Owners []string `json:"owners"`
}

// UpdateSourceRequest represents the update request structure for
// source.
type UpdateSourceRequest struct {
	// ID is the id of the source.
	ID                  uint `json:"id" binding:"required"`
	CreateSourceRequest `json:",inline" yaml:",inline"`
}

func (payload *CreateSourceRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *UpdateSourceRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

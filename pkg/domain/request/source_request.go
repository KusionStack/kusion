package request

import (
	"net/http"

	"kusionstack.io/kusion/pkg/domain/constant"
)

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
	ID uint `json:"id" binding:"required"`
	// Name is the name of the source.
	Name string `json:"name"`
	// SourceProvider is the type of the source provider.
	SourceProvider string `json:"sourceProvider"`
	// Remote is the source URL, including scheme.
	Remote string `json:"remote"`
	// Description is a human-readable description of the source.
	Description string `json:"description"`
	// Labels are custom labels associated with the source.
	Labels []string `json:"labels"`
	// Owners is a list of owners for the source.
	Owners []string `json:"owners"`
}

func (payload *CreateSourceRequest) Validate() error {
	// Validate source name
	if payload.Name == "" {
		return constant.ErrEmptySourceName
	}

	if validName(payload.Name) {
		return constant.ErrInvalidSourceName
	}

	// Validate source provider
	if payload.SourceProvider == "" {
		return constant.ErrEmptySourceProvider
	}

	if payload.SourceProvider != string(constant.SourceProviderTypeGit) &&
		payload.SourceProvider != string(constant.SourceProviderTypeGithub) &&
		payload.SourceProvider != string(constant.SourceProviderTypeOCI) &&
		payload.SourceProvider != string(constant.SourceProviderTypeLocal) {
		return constant.ErrInvalidSourceProvider
	}

	// Validate source remote is a valid URL
	if payload.Remote == "" {
		return constant.ErrEmptySourceRemote
	}

	if err := validURL(payload.Remote); err != nil {
		return err
	}

	return nil
}

func (payload *UpdateSourceRequest) Validate() error {
	// Validate source name
	if payload.Name != "" && validName(payload.Name) {
		return constant.ErrInvalidSourceName
	}

	// Validate source provider
	if payload.SourceProvider != "" &&
		payload.SourceProvider != string(constant.SourceProviderTypeGit) &&
		payload.SourceProvider != string(constant.SourceProviderTypeGithub) &&
		payload.SourceProvider != string(constant.SourceProviderTypeOCI) &&
		payload.SourceProvider != string(constant.SourceProviderTypeLocal) {
		return constant.ErrInvalidSourceProvider
	}

	// Validate source remote is a valid URL
	if payload.Remote == "" {
		return constant.ErrEmptySourceRemote
	}

	if payload.Remote != "" {
		if err := validURL(payload.Remote); err != nil {
			return err
		}
	}

	return nil
}

func (payload *CreateSourceRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *UpdateSourceRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

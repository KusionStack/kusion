package vclient

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// SecretList defines model for SecretList.
type SecretList struct {
	CreatedAt *time.Time `json:"created_at,omitempty"`
	ID        *uuid.UUID `json:"id,omitempty"`
	Name      string     `json:"name,omitempty"`
}

// SecretManagerSecretsListParams defines parameters for SecretManagerSecretsList.
type SecretManagerSecretsListParams struct {
	Name *string `form:"name,omitempty" json:"name,omitempty"`

	// Page A page number within the paginated result set.
	Page *int `form:"page,omitempty" json:"page,omitempty"`

	// PageSize Number of results to return per page.
	PageSize *int `form:"page_size,omitempty" json:"page_size,omitempty"`

	// ProjectID The project id.
	ProjectID uuid.UUID `json:"project-id,omitempty"`
}

// PaginatedSecretListList defines model for PaginatedSecretListList.
type PaginatedSecretListList struct {
	Count    *int          `json:"count,omitempty"`
	Next     *string       `json:"next,omitempty"`
	Previous *string       `json:"previous,omitempty"`
	Results  *[]SecretList `json:"results,omitempty"`
}

// SecretRetrieve defines model for SecretRetrieve.
type SecretRetrieve struct {
	CreatedAt *time.Time              `json:"created_at,omitempty"`
	ID        *uuid.UUID              `json:"id,omitempty"`
	Metadata  *map[string]interface{} `json:"metadata,omitempty"`
	Name      string                  `json:"name,omitempty"`
	Secret    *map[string]interface{} `json:"secret,omitempty"`
}

// SecretManagerSecretsRetrieveParams defines parameters for SecretManagerSecretsRetrieve.
type SecretManagerSecretsRetrieveParams struct {
	// ProjectID The project id.
	ProjectID uuid.UUID `json:"project-id,omitempty"`
}

// ErrorResponse defines model for ErrorResponse.
type ErrorResponse struct {
	Union json.RawMessage
}

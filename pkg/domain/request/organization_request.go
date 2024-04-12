package request

// CreateOrganizationRequest represents the create request structure for
// organization.
type CreateOrganizationRequest struct {
	// Name is the name of the organization.
	Name string `json:"name" binding:"required"`
	// Description is a human-readable description of the organization.
	Description string `json:"description"`
	// Labels are custom labels associated with the organization.
	Labels []string `json:"labels"`
	// Owners is a list of owners for the organization.
	Owners []string `json:"owners" binding:"required"`
}

// UpdateOrganizationRequest represents the update request structure for
// organization.
type UpdateOrganizationRequest struct {
	// ID is the id of the organization.
	ID uint `json:"id" binding:"required"`
	// Name is the name of the organization.
	Name string `json:"name"`
	// Description is a human-readable description of the organization.
	Description string `json:"description"`
	// Labels are custom labels associated with the organization.
	Labels map[string]string `json:"labels"`
	// Owners is a list of owners for the organization.
	Owners []string `json:"owners" binding:"required"`
}

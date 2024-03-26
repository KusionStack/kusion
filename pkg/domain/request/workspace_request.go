package request

// CreateWorkspaceRequest represents the create request structure for
// workspace.
type CreateWorkspaceRequest struct {
	// Name is the name of the workspace.
	Name string `json:"name" binding:"required"`
	// Description is a human-readable description of the workspace.
	Description string `json:"description"`
	// Labels are custom labels associated with the workspace.
	Labels []string `json:"labels"`
	// Owners is a list of owners for the workspace.
	Owners []string `json:"owners" binding:"required"`
	// BackendID is the configuration backend id associated with the workspace.
	BackendID uint `json:"backendID,string" binding:"required"`
}

// UpdateWorkspaceRequest represents the update request structure for
// workspace.
type UpdateWorkspaceRequest struct {
	// ID is the id of the workspace.
	ID uint `json:"id" binding:"required"`
	// Name is the name of the workspace.
	Name string `json:"name"`
	// Description is a human-readable description of the workspace.
	Description string `json:"description"`
	// Labels are custom labels associated with the workspace.
	Labels map[string]string `json:"labels"`
	// Owners is a list of owners for the workspace.
	Owners []string `json:"owners" binding:"required"`
	// BackendID is the configuration backend id associated with the workspace.
	BackendID uint `json:"backendID,string" binding:"required"`
}

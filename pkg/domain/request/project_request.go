package request

// CreateProjectRequest represents the create request structure for
// project.
type CreateProjectRequest struct {
	// Name is the name of the project.
	Name string `json:"name" binding:"required"`
	// SourceID is the configuration source id associated with the project.
	SourceID uint `json:"sourceID,string" binding:"required"`
	// OrganizationID is the organization id associated with the project.
	OrganizationID uint `json:"organizationID,string" binding:"required"`
	// Description is a human-readable description of the project.
	Description string `json:"description"`
	// Path is the relative path of the project within the sourcs..
	Path string `json:"path" binding:"required"`
	// Labels are custom labels associated with the project.
	Labels []string `json:"labels"`
	// Owners is a list of owners for the project.
	Owners []string `json:"owners"`
}

// UpdateProjectRequest represents the update request structure for
// project.
type UpdateProjectRequest struct {
	// ID is the id of the project.
	ID uint `json:"id" binding:"required"`
	// SourceID is the configuration source id associated with the project.
	SourceID uint `json:"sourceID,string"`
	// OrganizationID is the organization id associated with the project.
	OrganizationID uint `json:"organizationID,string"`
	// Name is the name of the project.
	Name string `json:"name"`
	// Description is a human-readable description of the project.
	Description string `json:"description"`
	// Path is the relative path of the project within the sourcs..
	Path string `json:"path"`
	// Labels are custom labels associated with the project.
	Labels map[string]string `json:"labels"`
	// Owners is a list of owners for the project.
	Owners []string `json:"owners"`
}

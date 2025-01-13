package response

import "kusionstack.io/kusion/pkg/domain/entity"

type PaginatedWorkspaceResponse struct {
	Workspaces  []*entity.Workspace `json:"workspaces"`
	Total       int                 `json:"total"`
	CurrentPage int                 `json:"currentPage"`
	PageSize    int                 `json:"pageSize"`
}

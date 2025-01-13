package response

import "kusionstack.io/kusion/pkg/domain/entity"

type PaginatedProjectResponse struct {
	Projects    []*entity.Project `json:"projects"`
	Total       int               `json:"total"`
	CurrentPage int               `json:"currentPage"`
	PageSize    int               `json:"pageSize"`
}

package response

import "kusionstack.io/kusion/pkg/domain/entity"

type PaginatedSourceResponse struct {
	Sources     []*entity.Source `json:"sources"`
	Total       int              `json:"total"`
	CurrentPage int              `json:"currentPage"`
	PageSize    int              `json:"pageSize"`
}

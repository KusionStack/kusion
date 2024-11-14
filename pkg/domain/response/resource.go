package response

import (
	"kusionstack.io/kusion/pkg/domain/entity"
)

type PaginatedResourceResponse struct {
	Resources   []*entity.Resource `json:"resources"`
	Total       int                `json:"total"`
	CurrentPage int                `json:"currentPage"`
	PageSize    int                `json:"pageSize"`
}

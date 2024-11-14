package response

import (
	"kusionstack.io/kusion/pkg/domain/entity"
)

type PaginatedRunResponse struct {
	Runs        []*entity.Run `json:"runs"`
	Total       int           `json:"total"`
	CurrentPage int           `json:"currentPage"`
	PageSize    int           `json:"pageSize"`
}

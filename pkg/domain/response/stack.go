package response

import "kusionstack.io/kusion/pkg/domain/entity"

type PaginatedStackResponse struct {
	Stacks      []*entity.Stack `json:"stacks"`
	Total       int             `json:"total"`
	CurrentPage int             `json:"currentPage"`
	PageSize    int             `json:"pageSize"`
}

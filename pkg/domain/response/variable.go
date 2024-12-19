package response

import "kusionstack.io/kusion/pkg/domain/entity"

type PaginatedVariableResponse struct {
	Variables   []*entity.Variable `json:"variables"`
	Total       int                `json:"total"`
	CurrentPage int                `json:"currentPage"`
	PageSize    int                `json:"pageSize"`
}

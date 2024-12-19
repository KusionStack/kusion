package response

import "kusionstack.io/kusion/pkg/domain/entity"

type PaginatedVariableLabelsResponse struct {
	VariableLabels []*entity.VariableLabels `json:"variableLabels"`
	Total          int                      `json:"total"`
	CurrentPage    int                      `json:"currentPage"`
	PageSize       int                      `json:"pageSize"`
}

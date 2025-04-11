package response

import "kusionstack.io/kusion/pkg/domain/entity"

type PaginatedVariableSetResponse struct {
	VariableSets []*entity.VariableSet `json:"variableSets"`
	Total        int                   `json:"total"`
	CurrentPage  int                   `json:"currentPage"`
	PageSize     int                   `json:"pageSize"`
}

type SelectedVariableSetResponse struct {
	VariableSets []*entity.VariableSet `json:"variableSets"`
	Total        int                   `json:"total"`
}

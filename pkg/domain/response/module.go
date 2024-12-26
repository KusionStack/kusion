package response

import "kusionstack.io/kusion/pkg/domain/entity"

type PaginatedModuleResponse struct {
	Modules            []*entity.Module            `json:"modules"`
	ModulesWithVersion []*entity.ModuleWithVersion `json:"modulesWithVersion"`
	Total              int                         `json:"total"`
	CurrentPage        int                         `json:"currentPage"`
	PageSize           int                         `json:"pageSize"`
}

package response

import "kusionstack.io/kusion/pkg/domain/entity"

type PaginatedBackendResponse struct {
	Backends    []*entity.Backend `json:"backends"`
	Total       int               `json:"total"`
	CurrentPage int               `json:"currentPage"`
	PageSize    int               `json:"pageSize"`
}

package response

import "kusionstack.io/kusion/pkg/domain/entity"

type PaginatedOrganizationResponse struct {
	Organizations []*entity.Organization `json:"organizations"`
	Total         int                    `json:"total"`
	CurrentPage   int                    `json:"currentPage"`
	PageSize      int                    `json:"pageSize"`
}

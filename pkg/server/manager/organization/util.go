package organization

import (
	"fmt"

	"kusionstack.io/kusion/pkg/domain/constant"
)

func validateOrganizationSortOptions(sortBy string) (string, error) {
	if sortBy == "" {
		return constant.SortByID, nil
	}
	if sortBy != constant.SortByID && sortBy != constant.SortByName && sortBy != constant.SortByCreateTimestamp {
		return "", fmt.Errorf("invalid sort option: %s. Can only sort by id or create timestamp", sortBy)
	}
	switch sortBy {
	case constant.SortByCreateTimestamp:
		return "created_at", nil
	case constant.SortByModifiedTimestamp:
		return "updated_at", nil
	}
	return sortBy, nil
}

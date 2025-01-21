package module

import (
	"fmt"

	"kusionstack.io/kusion/pkg/domain/constant"
)

func validateModuleSortOptions(sortBy string) (string, error) {
	if sortBy == "" {
		return constant.SortByID, nil
	}
	if sortBy != constant.SortByID && sortBy != constant.SortByName {
		return "", fmt.Errorf("invalid sort option: %s. Can only sort by id, name", sortBy)
	}
	return sortBy, nil
}

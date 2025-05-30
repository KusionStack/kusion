package variableset

import (
	"fmt"

	"kusionstack.io/kusion/pkg/domain/constant"
)

func validateVariableSetSortOptions(sortBy string) (string, error) {
	if sortBy == "" {
		return constant.SortByID, nil
	}
	if sortBy != constant.SortByID && sortBy != constant.SortByName {
		return "", fmt.Errorf("invalid sort option: %s. Can only sort by id, name", sortBy)
	}
	return sortBy, nil
}

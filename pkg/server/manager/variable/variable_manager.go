package variable

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

func (v *VariableManager) CreateVariable(ctx context.Context,
	requestPayload request.CreateVariableRequest,
) (*entity.Variable, error) {
	// Convert request payload to the domain model.
	var createdEntity entity.Variable
	if err := copier.Copy(&createdEntity, &requestPayload); err != nil {
		return nil, err
	}

	// Create variable with repository.
	if err := v.variableRepo.Create(ctx, &createdEntity); err != nil {
		return nil, err
	}

	return &createdEntity, nil
}

func (v *VariableManager) DeleteVariableByNameAndVariableSet(ctx context.Context,
	name, variableSet string,
) error {
	if err := v.variableRepo.Delete(ctx, name, variableSet); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGettingNonExistingVariable
		}
		return err
	}

	return nil
}

func (v *VariableManager) UpdateVariableByNameAndVariableSet(ctx context.Context,
	name, variableSet string, requestPayload request.UpdateVariableRequest,
) (*entity.Variable, error) {
	// Convert request payload to domain model.
	var requestEntity entity.Variable
	if err := copier.Copy(&requestEntity, &requestPayload); err != nil {
		return nil, err
	}

	// Get the existing variable by name.
	updatedEntity, err := v.variableRepo.Get(ctx, name, variableSet)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUpdatingNonExistingVariable
		}

		return nil, err
	}

	// Overwrite non-zero values in request entity to existing entity.
	copier.CopyWithOption(updatedEntity, requestEntity, copier.Option{IgnoreEmpty: true})

	// Update variable with repository.
	if err = v.variableRepo.Update(ctx, updatedEntity); err != nil {
		return nil, err
	}

	return updatedEntity, nil
}

func (v *VariableManager) GetVariableByNameAndVariableSet(ctx context.Context,
	name, variableSet string,
) (*entity.Variable, error) {
	existingEntity, err := v.variableRepo.Get(ctx, name, variableSet)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingVariable
		}

		return nil, err
	}

	return existingEntity, nil
}

func (v *VariableManager) ListVariables(ctx context.Context,
	filter *entity.VariableFilter, sortOptions *entity.SortOptions,
) (*entity.VariableListResult, error) {
	variableEntities, err := v.variableRepo.List(ctx, filter, sortOptions)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingVariable
		}

		return nil, err
	}

	return variableEntities, nil
}

func (v *VariableManager) BuildVariableFilterAndSortOptions(ctx context.Context,
	query *url.Values,
) (*entity.VariableFilter, *entity.SortOptions, error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Building variable filter and sort options...")

	variableNameParam := query.Get("variableName")
	variableSetNameParam := query.Get("variableSetName")
	fetchAllParam, _ := strconv.ParseBool(query.Get("fetchAll"))

	filter := entity.VariableFilter{}
	if variableNameParam != "" {
		filter.Name = variableNameParam
	}
	if variableSetNameParam != "" {
		filter.VariableSet = variableSetNameParam
	}
	filter.FetchAll = fetchAllParam

	// Set pagination parameters.
	page, _ := strconv.Atoi(query.Get("page"))
	if page <= 0 {
		page = constant.CommonPageDefault
	}
	pageSize, _ := strconv.Atoi(query.Get("pageSize"))
	if pageSize <= 0 {
		pageSize = constant.CommonPageSizeDefault
	}
	filter.Pagination = &entity.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	// Build sort options
	sortBy := query.Get("sortBy")
	sortBy, err := validateVariableSortOptions(sortBy)
	if err != nil {
		return nil, nil, err
	}
	SortOrderAscending, _ := strconv.ParseBool(query.Get("ascending"))
	variableSetSortOptions := &entity.SortOptions{
		Field:     sortBy,
		Ascending: SortOrderAscending,
	}

	return &filter, variableSetSortOptions, nil
}

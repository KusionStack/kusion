package variableset

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

func (vs *VariableSetManager) CreateVariableSet(ctx context.Context,
	requestPayload request.CreateVariableSetRequest,
) (*entity.VariableSet, error) {
	// Convert request payload to the domain model.
	var createdEntity entity.VariableSet
	if err := copier.Copy(&createdEntity, &requestPayload); err != nil {
		return nil, err
	}

	// Create variable set with repository.
	if err := vs.variableSetRepo.Create(ctx, &createdEntity); err != nil {
		return nil, err
	}

	return &createdEntity, nil
}

func (vs *VariableSetManager) DeleteVariableSetByName(ctx context.Context, name string) error {
	if err := vs.variableSetRepo.Delete(ctx, name); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGettingNonExistingVariableSet
		}
		return err
	}

	return nil
}

func (vs *VariableSetManager) UpdateVariableSetByName(ctx context.Context,
	name string, requestPayload request.UpdateVariableSetRequest,
) (*entity.VariableSet, error) {
	// Convert request payload to domain model.
	var requestEntity entity.VariableSet
	if err := copier.Copy(&requestEntity, &requestPayload); err != nil {
		return nil, err
	}

	// Get the existing variable set by name.
	updatedEntity, err := vs.variableSetRepo.Get(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUpdatingNonExistingVariableSet
		}

		return nil, err
	}

	// Overwrite non-zero values in request entity to existing entity.
	copier.CopyWithOption(updatedEntity, requestEntity, copier.Option{IgnoreEmpty: true})

	// Update variable set with repository.
	if err = vs.variableSetRepo.Update(ctx, updatedEntity); err != nil {
		return nil, err
	}

	return updatedEntity, nil
}

func (vs *VariableSetManager) GetVariableSetByName(ctx context.Context, name string) (*entity.VariableSet, error) {
	existingEntity, err := vs.variableSetRepo.Get(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingVariableSet
		}

		return nil, err
	}

	return existingEntity, nil
}

func (vs *VariableSetManager) ListVariableSets(ctx context.Context,
	filter *entity.VariableSetFilter, sortOptions *entity.SortOptions,
) (*entity.VariableSetListResult, error) {
	variableSetEntities, err := vs.variableSetRepo.List(ctx, filter, sortOptions)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingVariableSet
		}

		return nil, err
	}

	return variableSetEntities, nil
}

func (vs *VariableSetManager) BuildVariableSetFilterAndSortOptions(ctx context.Context,
	query *url.Values,
) (*entity.VariableSetFilter, *entity.SortOptions, error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Building variable set filter and sort options...")

	variableSetNameParam := query.Get("variableSetName")
	fetchAllParam, _ := strconv.ParseBool(query.Get("fetchAll"))

	filter := entity.VariableSetFilter{}
	if variableSetNameParam != "" {
		filter.Name = variableSetNameParam
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
	sortBy, err := validateVariableSetSortOptions(sortBy)
	if err != nil {
		return nil, nil, err
	}
	sortOrderDescending, _ := strconv.ParseBool(query.Get("descending"))
	variableSetSortOptions := &entity.SortOptions{
		Field:      sortBy,
		Descending: sortOrderDescending,
	}

	return &filter, variableSetSortOptions, nil
}

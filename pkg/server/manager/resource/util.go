package resource

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

func (m *ResourceManager) BuildResourceFilterAndSortOptions(ctx context.Context, query *url.Values) (*entity.ResourceFilter, *entity.SortOptions, error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Building resource filter...")

	filter := entity.ResourceFilter{}

	orgIDParam := query.Get("orgID")
	projectIDParam := query.Get("projectID")
	stackIDParam := query.Get("stackID")
	resourcePlaneParam := query.Get("resourcePlane")
	resourceTypeParam := query.Get("resourceType")

	if orgIDParam != "" {
		orgID, err := strconv.Atoi(orgIDParam)
		if err != nil {
			return nil, nil, constant.ErrInvalidOrganizationID
		}
		filter.OrgID = uint(orgID)
	}
	if projectIDParam != "" {
		// if project id is present, use project id
		projectID, err := strconv.Atoi(projectIDParam)
		if err != nil {
			return nil, nil, constant.ErrInvalidProjectID
		}
		filter.ProjectID = uint(projectID)
	}
	if stackIDParam != "" {
		// if stack id is present, use stack id
		stackID, err := strconv.Atoi(stackIDParam)
		if err != nil {
			return nil, nil, constant.ErrInvalidStackID
		}
		filter.StackID = uint(stackID)
	}
	if resourcePlaneParam != "" {
		// if resource plane is present, use resource plane
		filter.ResourcePlane = resourcePlaneParam
	}
	if resourceTypeParam != "" {
		// if resource type is present, use resource type
		filter.ResourceType = resourceTypeParam
	}

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
	sortBy, err := validateResourceSortOptions(sortBy)
	if err != nil {
		return nil, nil, err
	}
	SortOrderAscending, _ := strconv.ParseBool(query.Get("ascending"))
	projectSortOptions := &entity.SortOptions{
		Field:     sortBy,
		Ascending: SortOrderAscending,
	}

	return &filter, projectSortOptions, nil
}

func (m *ResourceManager) BuildResourceGraphFilter(ctx context.Context, query *url.Values) (*entity.ResourceFilter, error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Building resource graph filter...")

	filter := entity.ResourceFilter{}

	stackIDParam := query.Get("stackID")

	if stackIDParam != "" {
		// if stack id is present, use stack id
		stackID, err := strconv.Atoi(stackIDParam)
		if err != nil {
			return nil, constant.ErrInvalidStackID
		}
		filter.StackID = uint(stackID)
	} else {
		return nil, errors.New("stackID is required")
	}

	return &filter, nil
}

func validateResourceSortOptions(sortBy string) (string, error) {
	if sortBy == "" {
		return constant.SortByID, nil
	}
	if sortBy != constant.SortByID && sortBy != constant.SortByResourceName && sortBy != constant.SortByResourceURN && sortBy != constant.SortByCreateTimestamp {
		return "", fmt.Errorf("invalid sort option: %s. Can only sort by id, resource name, resource urn or create timestamp", sortBy)
	}
	switch sortBy {
	case constant.SortByCreateTimestamp:
		return "created_at", nil
	case constant.SortByModifiedTimestamp:
		return "updated_at", nil
	case constant.SortByResourceName:
		return "resource_name", nil
	case constant.SortByResourceURN:
		return "resource_urn", nil
	}
	return sortBy, nil
}

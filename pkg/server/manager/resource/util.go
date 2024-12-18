package resource

import (
	"context"
	"net/url"
	"strconv"

	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

func (m *ResourceManager) BuildResourceFilter(ctx context.Context,
	orgIDParam, projectIDParam, stackIDParam, resourcePlaneParam, resourceTypeParam string,
	query *url.Values,
) (*entity.ResourceFilter, error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Building resource filter...")
	filter := entity.ResourceFilter{}
	if orgIDParam != "" {
		orgID, err := strconv.Atoi(orgIDParam)
		if err != nil {
			return nil, constant.ErrInvalidOrganizationID
		}
		filter.OrgID = uint(orgID)
	}
	if projectIDParam != "" {
		// if project id is present, use project id
		projectID, err := strconv.Atoi(projectIDParam)
		if err != nil {
			return nil, constant.ErrInvalidProjectID
		}
		filter.ProjectID = uint(projectID)
	}
	if stackIDParam != "" {
		// if stack id is present, use stack id
		stackID, err := strconv.Atoi(stackIDParam)
		if err != nil {
			return nil, constant.ErrInvalidStackID
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
		page = constant.RunPageDefault
	}
	pageSize, _ := strconv.Atoi(query.Get("pageSize"))
	if pageSize <= 0 {
		pageSize = constant.RunPageSizeDefault
	}
	filter.Pagination = &entity.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return &filter, nil
}

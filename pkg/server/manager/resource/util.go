package resource

import (
	"context"
	"net/url"
	"strconv"

	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

func (m *ResourceManager) BuildResourceFilter(ctx context.Context, query *url.Values) (*entity.ResourceFilter, error) {
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

	return &filter, nil
}

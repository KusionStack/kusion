package resource

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/render"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/response"
	"kusionstack.io/kusion/pkg/server/handler"
	resourcemanager "kusionstack.io/kusion/pkg/server/manager/resource"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

// @Id				listResource
// @Summary		List resource
// @Description	List resource information
// @Tags			resource
// @Produce		json
// @Param			orgID			query		uint														false	"The organization ID"
// @Param			projectID		query		uint														false	"The project ID"
// @Param			stackID			query		uint														false	"The stack ID"
// @Param			resourceType	query		string														false	"The resource type"
// @Param			resourcePlane	query		string														false	"The resource plane"
// @Param			page			query		uint														false	"The current page to fetch. Default to 1"
// @Param			pageSize		query		uint														false	"The size of the page. Default to 10"
// @Success		200				{object}	handler.Response{data=[]response.PaginatedResourceResponse}	"Success"
// @Failure		400				{object}	error														"Bad Request"
// @Failure		401				{object}	error														"Unauthorized"
// @Failure		429				{object}	error														"Too Many Requests"
// @Failure		404				{object}	error														"Not Found"
// @Failure		500				{object}	error														"Internal Server Error"
// @Router			/api/v1/resources [get]
func (h *Handler) ListResources() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Listing resource...")

		query := r.URL.Query()
		filter, err := h.resourceManager.BuildResourceFilter(ctx, &query)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// List resources
		resourceEntities, err := h.resourceManager.ListResources(ctx, filter)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		paginatedResponse := response.PaginatedResourceResponse{
			Resources:   resourceEntities.Resources,
			Total:       resourceEntities.Total,
			CurrentPage: filter.Pagination.Page,
			PageSize:    filter.Pagination.PageSize,
		}
		handler.HandleResult(w, r, ctx, err, paginatedResponse)
	}
}

// @Id				getResource
// @Summary		Get resource
// @Description	Get resource information by resource ID
// @Tags			resource
// @Produce		json
// @Param			id	path		int										true	"Resource ID"
// @Success		200	{object}	handler.Response{data=entity.Resource}	"Success"
// @Failure		400	{object}	error									"Bad Request"
// @Failure		401	{object}	error									"Unauthorized"
// @Failure		429	{object}	error									"Too Many Requests"
// @Failure		404	{object}	error									"Not Found"
// @Failure		500	{object}	error									"Internal Server Error"
// @Router			/api/v1/resources/{id} [get]
func (h *Handler) GetResource() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Getting resource...", "resourceID", params.ResourceID)

		existingEntity, err := h.resourceManager.GetResourceByID(ctx, params.ResourceID)
		handler.HandleResult(w, r, ctx, err, existingEntity)
	}
}

// @Id				getResourceGraph
// @Summary		Get resource graph
// @Description	Get resource graph by stack ID
// @Tags			resource
// @Produce		json
// @Param			stackID	query		uint										true	"Stack ID"
// @Success		200		{object}	handler.Response{data=entity.ResourceGraph}	"Success"
// @Failure		400		{object}	error										"Bad Request"
// @Failure		401		{object}	error										"Unauthorized"
// @Failure		429		{object}	error										"Too Many Requests"
// @Failure		404		{object}	error										"Not Found"
// @Failure		500		{object}	error										"Internal Server Error"
// @Router			/api/v1/resources/graph [get]
func (h *Handler) GetResourceGraph() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Getting resource graph...")
		query := r.URL.Query()
		filter, err := h.resourceManager.BuildResourceGraphFilter(ctx, &query)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// List resources
		resourceEntities, err := h.resourceManager.ListResources(ctx, filter)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		resourceGraph := entity.NewResourceGraph()
		if err := resourceGraph.ConstructResourceGraph(resourceEntities.Resources); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		handler.HandleResult(w, r, ctx, nil, resourceGraph)
	}
}

func requestHelper(r *http.Request) (context.Context, *httplog.Logger, *ResourceRequestParams, error) {
	ctx := r.Context()
	resourceID := chi.URLParam(r, "resourceID")
	// Get resource with repository
	id, err := strconv.Atoi(resourceID)
	if err != nil {
		return nil, nil, nil, resourcemanager.ErrInvalidResourceID
	}
	logger := logutil.GetLogger(ctx)
	params := ResourceRequestParams{
		ResourceID: uint(id),
	}
	return ctx, logger, &params, nil
}

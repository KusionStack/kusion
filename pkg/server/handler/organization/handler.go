package organization

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/render"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/domain/response"
	"kusionstack.io/kusion/pkg/server/handler"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

// @Id				createOrganization
// @Summary		Create organization
// @Description	Create a new organization
// @Tags			organization
// @Accept			json
// @Produce		json
// @Param			organization	body		request.CreateOrganizationRequest			true	"Created organization"
// @Success		200				{object}	handler.Response{data=entity.Organization}	"Success"
// @Failure		400				{object}	error										"Bad Request"
// @Failure		401				{object}	error										"Unauthorized"
// @Failure		429				{object}	error										"Too Many Requests"
// @Failure		404				{object}	error										"Not Found"
// @Failure		500				{object}	error										"Internal Server Error"
// @Router			/api/v1/orgs [post]
func (h *Handler) CreateOrganization() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Creating organization...")

		// Decode the request body into the payload.
		var requestPayload request.CreateOrganizationRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Validate request payload
		if err := requestPayload.Validate(); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Create entity
		createdEntity, err := h.organizationManager.CreateOrganization(ctx, requestPayload)
		handler.HandleResult(w, r, ctx, err, createdEntity)
	}
}

// @Id				deleteOrganization
// @Summary		Delete organization
// @Description	Delete specified organization by ID
// @Tags			organization
// @Produce		json
// @Param			orgID	path		int								true	"Organization ID"
// @Success		200		{object}	handler.Response{data=string}	"Success"
// @Failure		400		{object}	error							"Bad Request"
// @Failure		401		{object}	error							"Unauthorized"
// @Failure		429		{object}	error							"Too Many Requests"
// @Failure		404		{object}	error							"Not Found"
// @Failure		500		{object}	error							"Internal Server Error"
// @Router			/api/v1/orgs/{orgID} [delete]
func (h *Handler) DeleteOrganization() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Deleting organization...")

		err = h.organizationManager.DeleteOrganizationByID(ctx, params.OrganizationID)
		handler.HandleResult(w, r, ctx, err, "Deletion Success")
	}
}

// @Id				updateOrganization
// @Summary		Update organization
// @Description	Update the specified organization
// @Tags			organization
// @Accept			json
// @Produce		json
// @Param			orgID			path		int											true	"Organization ID"
// @Param			organization	body		request.UpdateOrganizationRequest			true	"Updated organization"
// @Success		200				{object}	handler.Response{data=entity.Organization}	"Success"
// @Failure		400				{object}	error										"Bad Request"
// @Failure		401				{object}	error										"Unauthorized"
// @Failure		429				{object}	error										"Too Many Requests"
// @Failure		404				{object}	error										"Not Found"
// @Failure		500				{object}	error										"Internal Server Error"
// @Router			/api/v1/orgs/{orgID} [put]
func (h *Handler) UpdateOrganization() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Updating organization...")

		// Decode the request body into the payload.
		var requestPayload request.UpdateOrganizationRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Validate request payload
		if err := requestPayload.Validate(); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Update entity
		updatedEntity, err := h.organizationManager.UpdateOrganizationByID(ctx, params.OrganizationID, requestPayload)
		handler.HandleResult(w, r, ctx, err, updatedEntity)
	}
}

// @Id				getOrganization
// @Summary		Get organization
// @Description	Get organization information by organization ID
// @Tags			organization
// @Produce		json
// @Param			orgID	path		int											true	"Organization ID"
// @Success		200		{object}	handler.Response{data=entity.Organization}	"Success"
// @Failure		400		{object}	error										"Bad Request"
// @Failure		401		{object}	error										"Unauthorized"
// @Failure		429		{object}	error										"Too Many Requests"
// @Failure		404		{object}	error										"Not Found"
// @Failure		500		{object}	error										"Internal Server Error"
// @Router			/api/v1/orgs/{orgID} [get]
func (h *Handler) GetOrganization() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Getting organization...")

		existingEntity, err := h.organizationManager.GetOrganizationByID(ctx, params.OrganizationID)
		handler.HandleResult(w, r, ctx, err, existingEntity)
	}
}

// @Id				listOrganization
// @Summary		List organizations
// @Description	List all organizations
// @Tags			organization
// @Produce		json
// @Param			page		query		uint															false	"The current page to fetch. Default to 1"
// @Param			pageSize	query		uint															false	"The size of the page. Default to 10"
// @Param			sortBy		query		string														false	"Which field to sort the list by. Default to id"
// @Param			ascending	query		bool														false	"Whether to sort the list in ascending order. Default to false"
// @Success		200			{object}	handler.Response{data=response.PaginatedOrganizationResponse}	"Success"
// @Failure		400			{object}	error															"Bad Request"
// @Failure		401			{object}	error															"Unauthorized"
// @Failure		429			{object}	error															"Too Many Requests"
// @Failure		404			{object}	error															"Not Found"
// @Failure		500			{object}	error															"Internal Server Error"
// @Router			/api/v1/orgs [get]
func (h *Handler) ListOrganizations() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Listing organization...")

		query := r.URL.Query()
		filter, organizationSortOptions, err := h.organizationManager.BuildOrganizationFilterAndSortOptions(ctx, &query)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// List organizations with pagination.
		organizationEntities, err := h.organizationManager.ListOrganizations(ctx, filter, organizationSortOptions)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		paginatedResponse := response.PaginatedOrganizationResponse{
			Organizations: organizationEntities.Organizations,
			Total:         organizationEntities.Total,
			CurrentPage:   filter.Pagination.Page,
			PageSize:      filter.Pagination.PageSize,
		}
		handler.HandleResult(w, r, ctx, err, paginatedResponse)
	}
}

func requestHelper(r *http.Request) (context.Context, *httplog.Logger, *OrganizationRequestParams, error) {
	ctx := r.Context()
	organizationID := chi.URLParam(r, "organizationID")
	// Get stack with repository
	id, err := strconv.Atoi(organizationID)
	if err != nil {
		return nil, nil, nil, constant.ErrInvalidOrganizationID
	}
	logger := logutil.GetLogger(ctx)
	params := OrganizationRequestParams{
		OrganizationID: uint(id),
	}
	return ctx, logger, &params, nil
}

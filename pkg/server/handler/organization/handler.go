package organization

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-logr/logr"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/server/handler"
	organizationmanager "kusionstack.io/kusion/pkg/server/manager/organization"
	"kusionstack.io/kusion/pkg/server/util"
)

// @Summary      Create organization
// @Description  Create a new organization
// @Accept       json
// @Produce      json
// @Param        organization  body      CreateOrganizationRequest  true  "Created organization"
// @Success      200        {object}  entity.Organization        "Success"
// @Failure      400        {object}  errors.DetailError      "Bad Request"
// @Failure      401        {object}  errors.DetailError      "Unauthorized"
// @Failure      429        {object}  errors.DetailError      "Too Many Requests"
// @Failure      404        {object}  errors.DetailError      "Not Found"
// @Failure      500        {object}  errors.DetailError      "Internal Server Error"
// @Router       /api/v1/organization/{organizationName} [post]
func (h *Handler) CreateOrganization() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Creating organization...")

		// Decode the request body into the payload.
		var requestPayload request.CreateOrganizationRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		createdEntity, err := h.organizationManager.CreateOrganization(ctx, requestPayload)
		handler.HandleResult(w, r, ctx, err, createdEntity)
	}
}

// @Summary      Delete organization
// @Description  Delete specified organization by ID
// @Produce      json
// @Param        id   path      int                 true  "Organization ID"
// @Success      200  {object}  entity.Organization       "Success"
// @Failure      400             {object}  errors.DetailError   "Bad Request"
// @Failure      401             {object}  errors.DetailError   "Unauthorized"
// @Failure      429             {object}  errors.DetailError   "Too Many Requests"
// @Failure      404             {object}  errors.DetailError   "Not Found"
// @Failure      500             {object}  errors.DetailError   "Internal Server Error"
// @Router       /api/v1/organization/{organizationID} [delete]
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

// @Summary      Update organization
// @Description  Update the specified organization
// @Accept       json
// @Produce      json
// @Param        organization  body      UpdateOrganizationRequest  true  "Updated organization"
// @Success      200     {object}  entity.Organization        "Success"
// @Failure      400     {object}  errors.DetailError   "Bad Request"
// @Failure      401     {object}  errors.DetailError   "Unauthorized"
// @Failure      429     {object}  errors.DetailError   "Too Many Requests"
// @Failure      404     {object}  errors.DetailError   "Not Found"
// @Failure      500     {object}  errors.DetailError   "Internal Server Error"
// @Router       /api/v1/organization/{organizationID} [put]
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

		updatedEntity, err := h.organizationManager.UpdateOrganizationByID(ctx, params.OrganizationID, requestPayload)
		handler.HandleResult(w, r, ctx, err, updatedEntity)
	}
}

// @Summary      Get organization
// @Description  Get organization information by organization ID
// @Produce      json
// @Param        id   path      int                 true  "Organization ID"
// @Success      200  {object}  entity.Organization       "Success"
// @Failure      400  {object}  errors.DetailError  "Bad Request"
// @Failure      401  {object}  errors.DetailError  "Unauthorized"
// @Failure      429  {object}  errors.DetailError  "Too Many Requests"
// @Failure      404  {object}  errors.DetailError  "Not Found"
// @Failure      500  {object}  errors.DetailError  "Internal Server Error"
// @Router       /api/v1/organization/{organizationID} [get]
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

// @Summary      List organizations
// @Description  List all organizations
// @Produce      json
// @Success      200  {object}  entity.Organization       "Success"
// @Failure      400  {object}  errors.DetailError  "Bad Request"
// @Failure      401  {object}  errors.DetailError  "Unauthorized"
// @Failure      429  {object}  errors.DetailError  "Too Many Requests"
// @Failure      404  {object}  errors.DetailError  "Not Found"
// @Failure      500  {object}  errors.DetailError  "Internal Server Error"
// @Router       /api/v1/organization [get]
func (h *Handler) ListOrganizations() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Listing organization...")

		organizationEntities, err := h.organizationManager.ListOrganizations(ctx)
		handler.HandleResult(w, r, ctx, err, organizationEntities)
	}
}

func requestHelper(r *http.Request) (context.Context, *logr.Logger, *OrganizationRequestParams, error) {
	ctx := r.Context()
	organizationID := chi.URLParam(r, "organizationID")
	// Get stack with repository
	id, err := strconv.Atoi(organizationID)
	if err != nil {
		return nil, nil, nil, organizationmanager.ErrInvalidOrganizationID
	}
	logger := util.GetLogger(ctx)
	params := OrganizationRequestParams{
		OrganizationID: uint(id),
	}
	return ctx, &logger, &params, nil
}

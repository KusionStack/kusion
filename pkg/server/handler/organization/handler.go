package organization

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/server/handler"
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

		// Convert request payload to domain model
		var createdEntity entity.Organization
		if err := copier.Copy(&createdEntity, &requestPayload); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		// The default state is UnSynced
		createdEntity.CreationTimestamp = time.Now()
		createdEntity.UpdateTimestamp = time.Now()

		// Create organization with repository
		err := h.organizationRepo.Create(ctx, &createdEntity)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
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
// @Router       /api/v1/organization/{organizationName}  [delete]
// @Router       /api/v1/organization/{organizationID} [delete]
func (h *Handler) DeleteOrganization() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Deleting source...")
		organizationID := chi.URLParam(r, "organizationID")

		// Delete organization with repository
		id, err := strconv.Atoi(organizationID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidOrganizationID))
			return
		}
		err = h.organizationRepo.Delete(ctx, uint(id))
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
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
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Updating organization...")
		organizationID := chi.URLParam(r, "organizationID")

		// convert organization ID to int
		id, err := strconv.Atoi(organizationID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidOrganizationID))
			return
		}

		// Decode the request body into the payload.
		var requestPayload request.UpdateOrganizationRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Convert request payload to domain model
		var requestEntity entity.Organization
		if err := copier.Copy(&requestEntity, &requestPayload); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Get the existing organization by id
		updatedEntity, err := h.organizationRepo.Get(ctx, uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrUpdatingNonExistingOrganization))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Overwrite non-zero values in request entity to existing entity
		copier.CopyWithOption(updatedEntity, requestEntity, copier.Option{IgnoreEmpty: true})

		// Update organization with repository
		err = h.organizationRepo.Update(ctx, updatedEntity)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return updated organization
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
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Getting organization...")
		organizationID := chi.URLParam(r, "organizationID")

		// Get organization with repository
		id, err := strconv.Atoi(organizationID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidOrganizationID))
			return
		}
		existingEntity, err := h.organizationRepo.Get(ctx, uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrGettingNonExistingOrganization))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return found organization
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

		organizationEntities, err := h.organizationRepo.List(ctx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrGettingNonExistingOrganization))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return found organizations
		handler.HandleResult(w, r, ctx, err, organizationEntities)
	}
}

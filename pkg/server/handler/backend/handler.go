package backend

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

// @Summary      Create backend
// @Description  Create a new backend
// @Accept       json
// @Produce      json
// @Param        backend  body      CreateBackendRequest  true  "Created backend"
// @Success      200        {object}  entity.Backend        "Success"
// @Failure      400        {object}  errors.DetailError      "Bad Request"
// @Failure      401        {object}  errors.DetailError      "Unauthorized"
// @Failure      429        {object}  errors.DetailError      "Too Many Requests"
// @Failure      404        {object}  errors.DetailError      "Not Found"
// @Failure      500        {object}  errors.DetailError      "Internal Server Error"
// @Router       /api/v1/backend/{backendName} [post]
func (h *Handler) CreateBackend() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Creating backend...")

		// Decode the request body into the payload.
		var requestPayload request.CreateBackendRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Convert request payload to domain model
		var createdEntity entity.Backend
		if err := copier.Copy(&createdEntity, &requestPayload); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		// The default state is UnSynced
		createdEntity.CreationTimestamp = time.Now()
		createdEntity.UpdateTimestamp = time.Now()

		// Create backend with repository
		err := h.backendRepo.Create(ctx, &createdEntity)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		handler.HandleResult(w, r, ctx, err, createdEntity)
	}
}

// @Summary      Delete backend
// @Description  Delete specified backend by ID
// @Produce      json
// @Param        id   path      int                 true  "Backend ID"
// @Success      200  {object}  entity.Backend       "Success"
// @Failure      400             {object}  errors.DetailError   "Bad Request"
// @Failure      401             {object}  errors.DetailError   "Unauthorized"
// @Failure      429             {object}  errors.DetailError   "Too Many Requests"
// @Failure      404             {object}  errors.DetailError   "Not Found"
// @Failure      500             {object}  errors.DetailError   "Internal Server Error"
// @Router       /api/v1/backend/{backendName}  [delete]
// @Router       /api/v1/backend/{backendID} [delete]
func (h *Handler) DeleteBackend() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Deleting source...")
		backendID := chi.URLParam(r, "backendID")

		// Delete backend with repository
		id, err := strconv.Atoi(backendID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidBackendID))
			return
		}
		err = h.backendRepo.Delete(ctx, uint(id))
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		handler.HandleResult(w, r, ctx, err, "Deletion Success")
	}
}

// @Summary      Update backend
// @Description  Update the specified backend
// @Accept       json
// @Produce      json
// @Param        backend  body      UpdateBackendRequest  true  "Updated backend"
// @Success      200     {object}  entity.Backend        "Success"
// @Failure      400     {object}  errors.DetailError   "Bad Request"
// @Failure      401     {object}  errors.DetailError   "Unauthorized"
// @Failure      429     {object}  errors.DetailError   "Too Many Requests"
// @Failure      404     {object}  errors.DetailError   "Not Found"
// @Failure      500     {object}  errors.DetailError   "Internal Server Error"
// @Router       /api/v1/backend/{backendID} [put]
func (h *Handler) UpdateBackend() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Updating backend...")
		backendID := chi.URLParam(r, "backendID")

		// convert backend ID to int
		id, err := strconv.Atoi(backendID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidBackendID))
			return
		}

		// Decode the request body into the payload.
		var requestPayload request.UpdateBackendRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Convert request payload to domain model
		var requestEntity entity.Backend
		if err := copier.Copy(&requestEntity, &requestPayload); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Get the existing backend by id
		updatedEntity, err := h.backendRepo.Get(ctx, uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrUpdatingNonExistingBackend))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Overwrite non-zero values in request entity to existing entity
		copier.CopyWithOption(updatedEntity, requestEntity, copier.Option{IgnoreEmpty: true})

		// Update backend with repository
		err = h.backendRepo.Update(ctx, updatedEntity)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return updated backend
		handler.HandleResult(w, r, ctx, err, updatedEntity)
	}
}

// @Summary      Get backend
// @Description  Get backend information by backend ID
// @Produce      json
// @Param        id   path      int                 true  "Backend ID"
// @Success      200  {object}  entity.Backend       "Success"
// @Failure      400  {object}  errors.DetailError  "Bad Request"
// @Failure      401  {object}  errors.DetailError  "Unauthorized"
// @Failure      429  {object}  errors.DetailError  "Too Many Requests"
// @Failure      404  {object}  errors.DetailError  "Not Found"
// @Failure      500  {object}  errors.DetailError  "Internal Server Error"
// @Router       /api/v1/backend/{backendID} [get]
func (h *Handler) GetBackend() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Getting backend...")
		backendID := chi.URLParam(r, "backendID")

		// Get backend with repository
		id, err := strconv.Atoi(backendID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidBackendID))
			return
		}
		existingEntity, err := h.backendRepo.Get(ctx, uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrGettingNonExistingBackend))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return found backend
		handler.HandleResult(w, r, ctx, err, existingEntity)
	}
}

// @Summary      List backends
// @Description  List all backends
// @Produce      json
// @Success      200  {object}  entity.Backend       "Success"
// @Failure      400  {object}  errors.DetailError  "Bad Request"
// @Failure      401  {object}  errors.DetailError  "Unauthorized"
// @Failure      429  {object}  errors.DetailError  "Too Many Requests"
// @Failure      404  {object}  errors.DetailError  "Not Found"
// @Failure      500  {object}  errors.DetailError  "Internal Server Error"
// @Router       /api/v1/backend [get]
func (h *Handler) ListBackends() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Listing backend...")

		backendEntities, err := h.backendRepo.List(ctx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrGettingNonExistingBackend))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return found backends
		handler.HandleResult(w, r, ctx, err, backendEntities)
	}
}

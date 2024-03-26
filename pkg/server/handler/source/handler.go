package source

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/server/handler"
	"kusionstack.io/kusion/pkg/server/util"
)

// @Summary      Create source
// @Description  Create a new source
// @Accept       json
// @Produce      json
// @Param        source  body      CreateSourceRequest  true  "Created source"
// @Success      200     {object}  entity.Source        "Success"
// @Failure      400     {object}  errors.DetailError   "Bad Request"
// @Failure      401     {object}  errors.DetailError   "Unauthorized"
// @Failure      429     {object}  errors.DetailError   "Too Many Requests"
// @Failure      404     {object}  errors.DetailError   "Not Found"
// @Failure      500     {object}  errors.DetailError   "Internal Server Error"
// @Router       /api/v1/sourceID [post]
func (h *Handler) CreateSource() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Creating source...")

		// Decode the request body into the payload.
		var requestPayload request.CreateSourceRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Convert request payload to domain model
		var createdEntity entity.Source
		if err := copier.Copy(&createdEntity, &requestPayload); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Convert Remote string to URL
		remote, err := url.Parse(requestPayload.Remote)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		createdEntity.Remote = remote

		// Create source with repository
		err = h.sourceRepo.Create(ctx, &createdEntity)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return created entity
		handler.HandleResult(w, r, ctx, err, createdEntity)
	}
}

// @Summary      Delete source
// @Description  Delete specified source by ID
// @Produce      json
// @Param        id   path      int                 true  "Source ID"
// @Success      200  {object}  entity.Source       "Success"
// @Failure      400             {object}  errors.DetailError   "Bad Request"
// @Failure      401             {object}  errors.DetailError   "Unauthorized"
// @Failure      429             {object}  errors.DetailError   "Too Many Requests"
// @Failure      404             {object}  errors.DetailError   "Not Found"
// @Failure      500             {object}  errors.DetailError   "Internal Server Error"
// @Router       /api/v1/source/{id} [delete]
func (h *Handler) DeleteSource() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Deleting source...")
		sourceID := chi.URLParam(r, "sourceID")

		// Delete source with repository
		id, err := strconv.Atoi(sourceID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidSourceID))
			return
		}
		err = h.sourceRepo.Delete(ctx, uint(id))
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		handler.HandleResult(w, r, ctx, err, "Deletion Success")
	}
}

// @Summary      Update source
// @Description  Update the specified source
// @Accept       json
// @Produce      json
// @Param        source  body      UpdateSourceRequest  true  "Updated source"
// @Success      200     {object}  entity.Source        "Success"
// @Failure      400     {object}  errors.DetailError   "Bad Request"
// @Failure      401     {object}  errors.DetailError   "Unauthorized"
// @Failure      429     {object}  errors.DetailError   "Too Many Requests"
// @Failure      404     {object}  errors.DetailError   "Not Found"
// @Failure      500     {object}  errors.DetailError   "Internal Server Error"
// @Router       /api/v1/source/{id} [put]
func (h *Handler) UpdateSource() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Updating source...")
		sourceID := chi.URLParam(r, "sourceID")

		// Convert sourceID to int
		id, err := strconv.Atoi(sourceID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidSourceID))
			return
		}

		// Decode the request body into the payload.
		var requestPayload request.UpdateSourceRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Convert request payload to domain model
		var requestEntity entity.Source
		if err := copier.Copy(&requestEntity, &requestPayload); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Convert Remote string to URL
		remote, err := url.Parse(requestPayload.Remote)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		requestEntity.Remote = remote

		// Get the existing source by id
		updatedEntity, err := h.sourceRepo.Get(ctx, uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrUpdatingNonExistingSource))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Overwrite non-zero values in request entity to existing entity
		copier.CopyWithOption(updatedEntity, requestEntity, copier.Option{IgnoreEmpty: true})

		// Update source with repository
		err = h.sourceRepo.Update(ctx, updatedEntity)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return updated source
		handler.HandleResult(w, r, ctx, err, updatedEntity)
	}
}

// @Summary      Get source
// @Description  Get source information by source ID
// @Produce      json
// @Param        id   path      int                 true  "Source ID"
// @Success      200  {object}  entity.Source       "Success"
// @Failure      400  {object}  errors.DetailError  "Bad Request"
// @Failure      401  {object}  errors.DetailError  "Unauthorized"
// @Failure      429  {object}  errors.DetailError  "Too Many Requests"
// @Failure      404  {object}  errors.DetailError  "Not Found"
// @Failure      500  {object}  errors.DetailError  "Internal Server Error"
// @Router       /api/v1/source/{sourceID} [get]
func (h *Handler) GetSource() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Getting source...")
		sourceID := chi.URLParam(r, "sourceID")

		// Get source with repository
		id, err := strconv.Atoi(sourceID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidSourceID))
			return
		}
		existingEntity, err := h.sourceRepo.Get(ctx, uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrGettingNonExistingSource))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return found source
		handler.HandleResult(w, r, ctx, err, existingEntity)
	}
}

// @Summary      List source
// @Description  List source information by source ID
// @Produce      json
// @Success      200  {object}  entity.Source       "Success"
// @Failure      400  {object}  errors.DetailError  "Bad Request"
// @Failure      401  {object}  errors.DetailError  "Unauthorized"
// @Failure      429  {object}  errors.DetailError  "Too Many Requests"
// @Failure      404  {object}  errors.DetailError  "Not Found"
// @Failure      500  {object}  errors.DetailError  "Internal Server Error"
// @Router       /api/v1/source [get]
func (h *Handler) ListSources() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Listing source...")

		existingEntity, err := h.sourceRepo.List(ctx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrGettingNonExistingSource))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return found source
		handler.HandleResult(w, r, ctx, err, existingEntity)
	}
}

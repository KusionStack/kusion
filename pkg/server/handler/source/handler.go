package source

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-logr/logr"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/server/handler"
	sourcemanager "kusionstack.io/kusion/pkg/server/manager/source"
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

		// Return created entity
		createdEntity, err := h.sourceManager.CreateSource(ctx, requestPayload)
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
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Deleting source...")

		err = h.sourceManager.DeleteSourceByID(ctx, params.SourceID)
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
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Updating source...")

		// Decode the request body into the payload.
		var requestPayload request.UpdateSourceRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return updated source
		updatedEntity, err := h.sourceManager.UpdateSourceByID(ctx, params.SourceID, requestPayload)
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
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Getting source...")

		existingEntity, err := h.sourceManager.GetSourceByID(ctx, params.SourceID)
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

		// List sources
		sourceEntities, err := h.sourceManager.ListSources(ctx)
		handler.HandleResult(w, r, ctx, err, sourceEntities)
	}
}

func requestHelper(r *http.Request) (context.Context, *logr.Logger, *SourceRequestParams, error) {
	ctx := r.Context()
	sourceID := chi.URLParam(r, "sourceID")
	// Get stack with repository
	id, err := strconv.Atoi(sourceID)
	if err != nil {
		return nil, nil, nil, sourcemanager.ErrInvalidSourceID
	}
	logger := util.GetLogger(ctx)
	params := SourceRequestParams{
		SourceID: uint(id),
	}
	return ctx, &logger, &params, nil
}

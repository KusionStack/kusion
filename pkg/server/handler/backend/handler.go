package backend

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-logr/logr"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/server/handler"
	backendmanager "kusionstack.io/kusion/pkg/server/manager/backend"
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

		createdEntity, err := h.backendManager.CreateBackend(ctx, requestPayload)
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
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Deleting backend...", "backendID", params.BackendID)

		err = h.backendManager.DeleteBackendByID(ctx, params.BackendID)
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
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Updating backend..., backendID", params.BackendID)

		// Decode the request body into the payload.
		var requestPayload request.UpdateBackendRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		updatedEntity, err := h.backendManager.UpdateBackendByID(ctx, params.BackendID, requestPayload)
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
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Getting backend...", "backendID", params.BackendID)

		existingEntity, err := h.backendManager.GetBackendByID(ctx, params.BackendID)
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

		backendEntities, err := h.backendManager.ListBackends(ctx)
		handler.HandleResult(w, r, ctx, err, backendEntities)
	}
}

func requestHelper(r *http.Request) (context.Context, *logr.Logger, *BackendRequestParams, error) {
	ctx := r.Context()
	backendID := chi.URLParam(r, "backendID")
	// Get stack with repository
	id, err := strconv.Atoi(backendID)
	if err != nil {
		return nil, nil, nil, backendmanager.ErrInvalidBackendID
	}
	logger := util.GetLogger(ctx)
	params := BackendRequestParams{
		BackendID: uint(id),
	}
	return ctx, &logger, &params, nil
}

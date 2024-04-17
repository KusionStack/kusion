package workspace

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-logr/logr"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/server/handler"
	workspacemanager "kusionstack.io/kusion/pkg/server/manager/workspace"
	"kusionstack.io/kusion/pkg/server/util"
)

// @Summary      Create workspace
// @Description  Create a new workspace
// @Accept       json
// @Produce      json
// @Param        workspace  body      CreateWorkspaceRequest  true  "Created workspace"
// @Success      200        {object}  entity.Workspace        "Success"
// @Failure      400        {object}  errors.DetailError      "Bad Request"
// @Failure      401        {object}  errors.DetailError      "Unauthorized"
// @Failure      429        {object}  errors.DetailError      "Too Many Requests"
// @Failure      404        {object}  errors.DetailError      "Not Found"
// @Failure      500        {object}  errors.DetailError      "Internal Server Error"
// @Router       /api/v1/workspace/{workspaceName} [post]
func (h *Handler) CreateWorkspace() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Creating workspace...")

		// Decode the request body into the payload.
		var requestPayload request.CreateWorkspaceRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		createdEntity, err := h.workspaceManager.CreateWorkspace(ctx, requestPayload)
		handler.HandleResult(w, r, ctx, err, createdEntity)
	}
}

// @Summary      Delete workspace
// @Description  Delete specified workspace by ID
// @Produce      json
// @Param        id   path      int                 true  "Workspace ID"
// @Success      200  {object}  entity.Workspace       "Success"
// @Failure      400             {object}  errors.DetailError   "Bad Request"
// @Failure      401             {object}  errors.DetailError   "Unauthorized"
// @Failure      429             {object}  errors.DetailError   "Too Many Requests"
// @Failure      404             {object}  errors.DetailError   "Not Found"
// @Failure      500             {object}  errors.DetailError   "Internal Server Error"
// @Router       /api/v1/workspace/{workspaceName}  [delete]
// @Router       /api/v1/workspace/{workspaceID} [delete]
func (h *Handler) DeleteWorkspace() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Deleting source...", "workspaceID", params.WorkspaceID)

		err = h.workspaceManager.DeleteWorkspaceByID(ctx, params.WorkspaceID)
		handler.HandleResult(w, r, ctx, err, "Deletion Success")
	}
}

// @Summary      Update workspace
// @Description  Update the specified workspace
// @Accept       json
// @Produce      json
// @Param        workspace  body      UpdateWorkspaceRequest  true  "Updated workspace"
// @Success      200     {object}  entity.Workspace        "Success"
// @Failure      400     {object}  errors.DetailError   "Bad Request"
// @Failure      401     {object}  errors.DetailError   "Unauthorized"
// @Failure      429     {object}  errors.DetailError   "Too Many Requests"
// @Failure      404     {object}  errors.DetailError   "Not Found"
// @Failure      500     {object}  errors.DetailError   "Internal Server Error"
// @Router       /api/v1/workspace/{workspaceID} [put]
func (h *Handler) UpdateWorkspace() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Updating workspace...", "workspaceID", params.WorkspaceID)

		// Decode the request body into the payload.
		var requestPayload request.UpdateWorkspaceRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		updatedEntity, err := h.workspaceManager.UpdateWorkspaceByID(ctx, params.WorkspaceID, requestPayload)
		handler.HandleResult(w, r, ctx, err, updatedEntity)
	}
}

// @Summary      Get workspace
// @Description  Get workspace information by workspace ID
// @Produce      json
// @Param        id   path      int                 true  "Workspace ID"
// @Success      200  {object}  entity.Workspace       "Success"
// @Failure      400  {object}  errors.DetailError  "Bad Request"
// @Failure      401  {object}  errors.DetailError  "Unauthorized"
// @Failure      429  {object}  errors.DetailError  "Too Many Requests"
// @Failure      404  {object}  errors.DetailError  "Not Found"
// @Failure      500  {object}  errors.DetailError  "Internal Server Error"
// @Router       /api/v1/workspace/{workspaceID} [get]
func (h *Handler) GetWorkspace() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Getting workspace...", "workspaceID", params.WorkspaceID)

		// Return found workspace
		existingEntity, err := h.workspaceManager.GetWorkspaceByID(ctx, params.WorkspaceID)
		handler.HandleResult(w, r, ctx, err, existingEntity)
	}
}

// @Summary      List workspaces
// @Description  List all workspaces
// @Produce      json
// @Success      200  {object}  entity.Workspace       "Success"
// @Failure      400  {object}  errors.DetailError  "Bad Request"
// @Failure      401  {object}  errors.DetailError  "Unauthorized"
// @Failure      429  {object}  errors.DetailError  "Too Many Requests"
// @Failure      404  {object}  errors.DetailError  "Not Found"
// @Failure      500  {object}  errors.DetailError  "Internal Server Error"
// @Router       /api/v1/workspace [get]
func (h *Handler) ListWorkspaces() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Listing workspace...")

		// Return found workspaces
		workspaceEntities, err := h.workspaceManager.ListWorkspaces(ctx)
		handler.HandleResult(w, r, ctx, err, workspaceEntities)
	}
}

func requestHelper(r *http.Request) (context.Context, *logr.Logger, *WorkspaceRequestParams, error) {
	ctx := r.Context()
	workspaceID := chi.URLParam(r, "workspaceID")
	// Get stack with repository
	id, err := strconv.Atoi(workspaceID)
	if err != nil {
		return nil, nil, nil, workspacemanager.ErrInvalidWorkspaceID
	}
	logger := util.GetLogger(ctx)
	params := WorkspaceRequestParams{
		WorkspaceID: uint(id),
	}
	return ctx, &logger, &params, nil
}

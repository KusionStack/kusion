package workspace

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/render"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/server/handler"
	backendmanager "kusionstack.io/kusion/pkg/server/manager/backend"
	workspacemanager "kusionstack.io/kusion/pkg/server/manager/workspace"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

// @Id				createWorkspace
// @Summary		Create workspace
// @Description	Create a new workspace
// @Tags			workspace
// @Accept			json
// @Produce		json
// @Param			workspace	body		request.CreateWorkspaceRequest			true	"Created workspace"
// @Success		200			{object}	handler.Response{data=entity.Workspace}	"Success"
// @Failure		400			{object}	error									"Bad Request"
// @Failure		401			{object}	error									"Unauthorized"
// @Failure		429			{object}	error									"Too Many Requests"
// @Failure		404			{object}	error									"Not Found"
// @Failure		500			{object}	error									"Internal Server Error"
// @Router			/api/v1/workspaces [post]
func (h *Handler) CreateWorkspace() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
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

// @Id				deleteWorkspace
// @Summary		Delete workspace
// @Description	Delete specified workspace by ID
// @Tags			workspace
// @Produce		json
// @Param			id	path		int								true	"Workspace ID"
// @Success		200	{object}	handler.Response{data=string}	"Success"
// @Failure		400	{object}	error							"Bad Request"
// @Failure		401	{object}	error							"Unauthorized"
// @Failure		429	{object}	error							"Too Many Requests"
// @Failure		404	{object}	error							"Not Found"
// @Failure		500	{object}	error							"Internal Server Error"
// @Router			/api/v1/workspaces/{id} [delete]
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

// @Id				updateWorkspace
// @Summary		Update workspace
// @Description	Update the specified workspace
// @Tags			workspace
// @Accept			json
// @Produce		json
// @Param			id			path		int										true	"Workspace ID"
// @Param			workspace	body		request.UpdateWorkspaceRequest			true	"Updated workspace"
// @Success		200			{object}	handler.Response{data=entity.Workspace}	"Success"
// @Failure		400			{object}	error									"Bad Request"
// @Failure		401			{object}	error									"Unauthorized"
// @Failure		429			{object}	error									"Too Many Requests"
// @Failure		404			{object}	error									"Not Found"
// @Failure		500			{object}	error									"Internal Server Error"
// @Router			/api/v1/workspaces/{id} [put]
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

// @Id				getWorkspace
// @Summary		Get workspace
// @Description	Get workspace information by workspace ID
// @Tags			workspace
// @Produce		json
// @Param			id	path		int										true	"Workspace ID"
// @Success		200	{object}	handler.Response{data=entity.Workspace}	"Success"
// @Failure		400	{object}	error									"Bad Request"
// @Failure		401	{object}	error									"Unauthorized"
// @Failure		429	{object}	error									"Too Many Requests"
// @Failure		404	{object}	error									"Not Found"
// @Failure		500	{object}	error									"Internal Server Error"
// @Router			/api/v1/workspaces/{id} [get]
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

// @Id				listWorkspace
// @Summary		List workspaces
// @Description	List all workspaces
// @Tags			workspace
// @Produce		json
// @Success		200	{object}	handler.Response{data=[]entity.Workspace}	"Success"
// @Failure		400	{object}	error										"Bad Request"
// @Failure		401	{object}	error										"Unauthorized"
// @Failure		429	{object}	error										"Too Many Requests"
// @Failure		404	{object}	error										"Not Found"
// @Failure		500	{object}	error										"Internal Server Error"
// @Router			/api/v1/workspaces [get]
func (h *Handler) ListWorkspaces() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Listing workspace...")

		filter := entity.WorkspaceFilter{}
		backendIDParam := r.URL.Query().Get("backendID")
		if backendIDParam != "" {
			backendID, err := strconv.Atoi(backendIDParam)
			if err != nil {
				render.Render(w, r, handler.FailureResponse(ctx, backendmanager.ErrInvalidBackendID))
				return
			}
			filter.BackendID = uint(backendID)
		}
		name := r.URL.Query().Get("name")
		if name != "" {
			filter.Name = name
		}

		// Return found workspaces
		workspaceEntities, err := h.workspaceManager.ListWorkspaces(ctx, &filter)
		handler.HandleResult(w, r, ctx, err, workspaceEntities)
	}
}

func requestHelper(r *http.Request) (context.Context, *httplog.Logger, *WorkspaceRequestParams, error) {
	ctx := r.Context()
	workspaceID := chi.URLParam(r, "workspaceID")
	// Get stack with repository
	id, err := strconv.Atoi(workspaceID)
	if err != nil {
		return nil, nil, nil, workspacemanager.ErrInvalidWorkspaceID
	}
	logger := logutil.GetLogger(ctx)
	params := WorkspaceRequestParams{
		WorkspaceID: uint(id),
	}
	return ctx, logger, &params, nil
}

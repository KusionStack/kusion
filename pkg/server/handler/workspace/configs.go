package workspace

import (
	"context"
	"net/http"

	"github.com/go-chi/render"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/server/handler"
)

// @Id				getWorkspaceConfigs
// @Summary		get workspace configurations
// @Description	Get configurations in the specified workspace
// @Tags			workspace
// @Accept			json
// @Produce		json
// @Param			workspaceID	path		int							true	"Workspace ID"
// @Success		200			{object}	request.WorkspaceConfigs	"Success"
// @Failure		400			{object}	error						"Bad Request"
// @Failure		401			{object}	error						"Unauthorized"
// @Failure		429			{object}	error						"Too Many Requests"
// @Failure		404			{object}	error						"Not Found"
// @Failure		500			{object}	error						"Internal Server Error"
// @Router			/api/v1/workspaces/{workspaceID}/configs [get]
func (h *Handler) GetWorkspaceConfigs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from the context.
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Getting workspace configs...", "workspaceID", params.WorkspaceID)

		wsConfigs, err := h.workspaceManager.GetWorkspaceConfigs(ctx, params.WorkspaceID)
		handler.HandleResult(w, r, ctx, err, wsConfigs)
	}
}

// @Id				validateWorkspaceConfigs
// @Summary		Validate workspace configurations
// @Description	Validate the configurations in the specified workspace
// @Tags			workspace
// @Accept			json
// @Produce		json
// @Param			workspace	body		request.WorkspaceConfigs	true	"Workspace configurations to be validated"
// @Success		200			{object}	request.WorkspaceConfigs	"Success"
// @Failure		400			{object}	error						"Bad Request"
// @Failure		401			{object}	error						"Unauthorized"
// @Failure		429			{object}	error						"Too Many Requests"
// @Failure		404			{object}	error						"Not Found"
// @Failure		500			{object}	error						"Internal Server Error"
// @Router			/api/v1/workspaces/configs/validate [post]
func (h *Handler) ValidateWorkspaceConfigs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode the request body into the payload.
		var requestPayload request.WorkspaceConfigs
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(context.Background(), err))
			return
		}

		wsConfigs, err := h.workspaceManager.ValidateWorkspaceConfigs(context.Background(), requestPayload)
		handler.HandleResult(w, r, context.Background(), err, wsConfigs)
	}
}

// @Id				updateWorkspaceConfigs
// @Summary		Update workspace configurations
// @Description	Update the configurations in the specified workspace
// @Tags			workspace
// @Accept			json
// @Produce		json
// @Param			workspaceID	path		int							true	"Workspace ID"
// @Param			workspace	body		request.WorkspaceConfigs	true	"Updated workspace configurations"
// @Success		200			{object}	request.WorkspaceConfigs	"Success"
// @Failure		400			{object}	error						"Bad Request"
// @Failure		401			{object}	error						"Unauthorized"
// @Failure		429			{object}	error						"Too Many Requests"
// @Failure		404			{object}	error						"Not Found"
// @Failure		500			{object}	error						"Internal Server Error"
// @Router			/api/v1/workspaces/{workspaceID}/configs [put]
func (h *Handler) UpdateWorkspaceConfigs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Updating workspace configurations...", "workspaceID", params.WorkspaceID)

		// Decode the request body into the payload.
		var requestPayload request.WorkspaceConfigs
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		wsConfigs, err := h.workspaceManager.UpdateWorkspaceConfigs(ctx, params.WorkspaceID, requestPayload)
		handler.HandleResult(w, r, ctx, err, wsConfigs)
	}
}

// @Id				createWorkspaceModDeps
// @Summary		Create the module dependencies of the workspace
// @Description	Create the module dependencies in kcl.mod of the specified workspace
// @Tags			workspace
// @Accept			json
// @Produce		plain
// @Param			workspaceID											path		int		true	"Workspace ID"
// @Success		200													{object}	string	"Success"
// @Failure		400													{object}	error	"Bad Request"
// @Failure		401													{object}	error	"Unauthorized"
// @Failure		429													{object}	error	"Too Many Requests"
// @Failure		404													{object}	error	"Not Found"
// @Failure		500													{object}	error	"Internal Server Error"
// @Router			/api/v1/workspaces/{workspaceID}/configs/mod-deps 																																					[post]
func (h *Handler) CreateWorkspaceModDeps() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Creating kusion module dependencies...", "workspaceID", params.WorkspaceID)

		deps, err := h.workspaceManager.CreateKCLModDependencies(ctx, params.WorkspaceID)
		handler.HandleResult(w, r, ctx, err, deps)
	}
}

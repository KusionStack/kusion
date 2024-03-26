package workspace

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

		// Convert request payload to domain model
		var createdEntity entity.Workspace
		if err := copier.Copy(&createdEntity, &requestPayload); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		// The default state is UnSynced
		createdEntity.CreationTimestamp = time.Now()
		createdEntity.UpdateTimestamp = time.Now()

		// Get backend by id
		backendEntity, err := h.backendRepo.Get(ctx, requestPayload.BackendID)
		if err != nil && err == gorm.ErrRecordNotFound {
			render.Render(w, r, handler.FailureResponse(ctx, ErrBackendNotFound))
			return
		} else if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		createdEntity.Backend = backendEntity

		// Create workspace with repository
		err = h.workspaceRepo.Create(ctx, &createdEntity)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
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
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Deleting source...")
		workspaceID := chi.URLParam(r, "workspaceID")

		// Delete workspace with repository
		id, err := strconv.Atoi(workspaceID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidWorkspaceID))
			return
		}
		err = h.workspaceRepo.Delete(ctx, uint(id))
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
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
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Updating workspace...")
		workspaceID := chi.URLParam(r, "workspaceID")

		// convert workspace ID to int
		id, err := strconv.Atoi(workspaceID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidWorkspaceID))
			return
		}

		// Decode the request body into the payload.
		var requestPayload request.UpdateWorkspaceRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Convert request payload to domain model
		var requestEntity entity.Workspace
		if err := copier.Copy(&requestEntity, &requestPayload); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Get the existing workspace by id
		updatedEntity, err := h.workspaceRepo.Get(ctx, uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrUpdatingNonExistingWorkspace))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Overwrite non-zero values in request entity to existing entity
		copier.CopyWithOption(updatedEntity, requestEntity, copier.Option{IgnoreEmpty: true})

		// Update workspace with repository
		err = h.workspaceRepo.Update(ctx, updatedEntity)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return updated workspace
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
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Getting workspace...")
		workspaceID := chi.URLParam(r, "workspaceID")

		// Get workspace with repository
		id, err := strconv.Atoi(workspaceID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidWorkspaceID))
			return
		}
		existingEntity, err := h.workspaceRepo.Get(ctx, uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrGettingNonExistingWorkspace))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return found workspace
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

		workspaceEntities, err := h.workspaceRepo.List(ctx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrGettingNonExistingWorkspace))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return found workspaces
		handler.HandleResult(w, r, ctx, err, workspaceEntities)
	}
}

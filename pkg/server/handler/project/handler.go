package project

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-logr/logr"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/server/handler"
	projectmanager "kusionstack.io/kusion/pkg/server/manager/project"
	"kusionstack.io/kusion/pkg/server/util"
)

// @Summary      Create project
// @Description  Create a new project
// @Accept       json
// @Produce      json
// @Param        project  body      CreateProjectRequest  true  "Created project"
// @Success      200        {object}  entity.Project        "Success"
// @Failure      400        {object}  errors.DetailError      "Bad Request"
// @Failure      401        {object}  errors.DetailError      "Unauthorized"
// @Failure      429        {object}  errors.DetailError      "Too Many Requests"
// @Failure      404        {object}  errors.DetailError      "Not Found"
// @Failure      500        {object}  errors.DetailError      "Internal Server Error"
// @Router       /api/v1/project/{projectName} [post]
func (h *Handler) CreateProject() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Creating project...")

		// Decode the request body into the payload.
		var requestPayload request.CreateProjectRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		createdEntity, err := h.projectManager.CreateProject(ctx, requestPayload)
		handler.HandleResult(w, r, ctx, err, createdEntity)
	}
}

// @Summary      Delete project
// @Description  Delete specified project by ID
// @Produce      json
// @Param        id   path      int                 true  "Project ID"
// @Success      200  {object}  entity.Project       "Success"
// @Failure      400             {object}  errors.DetailError   "Bad Request"
// @Failure      401             {object}  errors.DetailError   "Unauthorized"
// @Failure      429             {object}  errors.DetailError   "Too Many Requests"
// @Failure      404             {object}  errors.DetailError   "Not Found"
// @Failure      500             {object}  errors.DetailError   "Internal Server Error"
// @Router       /api/v1/project/{projectName}  [delete]
// @Router       /api/v1/project/{projectID} [delete]
func (h *Handler) DeleteProject() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Deleting source...", "projectID", params.ProjectID)

		err = h.projectManager.DeleteProjectByID(ctx, params.ProjectID)
		handler.HandleResult(w, r, ctx, err, "Deletion Success")
	}
}

// @Summary      Update project
// @Description  Update the specified project
// @Accept       json
// @Produce      json
// @Param        project  body      UpdateProjectRequest  true  "Updated project"
// @Success      200     {object}  entity.Project        "Success"
// @Failure      400     {object}  errors.DetailError   "Bad Request"
// @Failure      401     {object}  errors.DetailError   "Unauthorized"
// @Failure      429     {object}  errors.DetailError   "Too Many Requests"
// @Failure      404     {object}  errors.DetailError   "Not Found"
// @Failure      500     {object}  errors.DetailError   "Internal Server Error"
// @Router       /api/v1/project/{projectID} [put]
func (h *Handler) UpdateProject() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Updating project...", "projectID", params.ProjectID)

		// Decode the request body into the payload.
		var requestPayload request.UpdateProjectRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		updatedEntity, err := h.projectManager.UpdateProjectByID(ctx, params.ProjectID, requestPayload)
		handler.HandleResult(w, r, ctx, err, updatedEntity)
	}
}

// @Summary      Get project
// @Description  Get project information by project ID
// @Produce      json
// @Param        id   path      int                 true  "Project ID"
// @Success      200  {object}  entity.Project       "Success"
// @Failure      400  {object}  errors.DetailError  "Bad Request"
// @Failure      401  {object}  errors.DetailError  "Unauthorized"
// @Failure      429  {object}  errors.DetailError  "Too Many Requests"
// @Failure      404  {object}  errors.DetailError  "Not Found"
// @Failure      500  {object}  errors.DetailError  "Internal Server Error"
// @Router       /api/v1/project/{projectID} [get]
func (h *Handler) GetProject() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Getting project...", "projectID", params.ProjectID)

		existingEntity, err := h.projectManager.GetProjectByID(ctx, params.ProjectID)
		handler.HandleResult(w, r, ctx, err, existingEntity)
	}
}

// @Summary      List projects
// @Description  List all projects
// @Produce      json
// @Success      200  {object}  entity.Project       "Success"
// @Failure      400  {object}  errors.DetailError  "Bad Request"
// @Failure      401  {object}  errors.DetailError  "Unauthorized"
// @Failure      429  {object}  errors.DetailError  "Too Many Requests"
// @Failure      404  {object}  errors.DetailError  "Not Found"
// @Failure      500  {object}  errors.DetailError  "Internal Server Error"
// @Router       /api/v1/project [get]
func (h *Handler) ListProjects() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Listing project...")

		projectEntities, err := h.projectManager.ListProjects(ctx)
		handler.HandleResult(w, r, ctx, err, projectEntities)
	}
}

func requestHelper(r *http.Request) (context.Context, *logr.Logger, *ProjectRequestParams, error) {
	ctx := r.Context()
	projectID := chi.URLParam(r, "projectID")
	// Get stack with repository
	id, err := strconv.Atoi(projectID)
	if err != nil {
		return nil, nil, nil, projectmanager.ErrInvalidProjectID
	}
	logger := util.GetLogger(ctx)
	params := ProjectRequestParams{
		ProjectID: uint(id),
	}
	return ctx, &logger, &params, nil
}

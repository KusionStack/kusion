package project

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/render"

	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/domain/response"
	"kusionstack.io/kusion/pkg/server/handler"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

// @Id				createProject
// @Summary		Create project
// @Description	Create a new project
// @Tags			project
// @Accept			json
// @Produce		json
// @Param			project	body		request.CreateProjectRequest			true	"Created project"
// @Success		200		{object}	handler.Response{data=entity.Project}	"Success"
// @Failure		400		{object}	error									"Bad Request"
// @Failure		401		{object}	error									"Unauthorized"
// @Failure		429		{object}	error									"Too Many Requests"
// @Failure		404		{object}	error									"Not Found"
// @Failure		500		{object}	error									"Internal Server Error"
// @Router			/api/v1/projects [post]
func (h *Handler) CreateProject() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Creating project...")

		// Decode the request body into the payload.
		var requestPayload request.CreateProjectRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Validate request payload
		if err := requestPayload.Validate(); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		createdEntity, err := h.projectManager.CreateProject(ctx, requestPayload)
		handler.HandleResult(w, r, ctx, err, createdEntity)

		defer func() {
			if err != nil {
				// Rollback
				err = h.projectManager.DeleteProjectByID(ctx, createdEntity.ID)
				if err != nil {
					logger.Info("Failed to rollback project creation", "projectID", createdEntity.ID, "error", err)
				}
			}
		}()
	}
}

// @Id				deleteProject
// @Summary		Delete project
// @Description	Delete specified project by ID
// @Tags			project
// @Produce		json
// @Param			projectID	path		int								true	"Project ID"
// @Success		200			{object}	handler.Response{data=string}	"Success"
// @Failure		400			{object}	error							"Bad Request"
// @Failure		401			{object}	error							"Unauthorized"
// @Failure		429			{object}	error							"Too Many Requests"
// @Failure		404			{object}	error							"Not Found"
// @Failure		500			{object}	error							"Internal Server Error"
// @Router			/api/v1/projects/{projectID} [delete]
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

// @Id				updateProject
// @Summary		Update project
// @Description	Update the specified project
// @Tags			project
// @Accept			json
// @Produce		json
// @Param			projectID	path		uint									true	"Project ID"
// @Param			project		body		request.UpdateProjectRequest			true	"Updated project"
// @Success		200			{object}	handler.Response{data=entity.Project}	"Success"
// @Failure		400			{object}	error									"Bad Request"
// @Failure		401			{object}	error									"Unauthorized"
// @Failure		429			{object}	error									"Too Many Requests"
// @Failure		404			{object}	error									"Not Found"
// @Failure		500			{object}	error									"Internal Server Error"
// @Router			/api/v1/projects/{projectID} [put]
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

		// Validate request payload
		if err := requestPayload.Validate(); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		updatedEntity, err := h.projectManager.UpdateProjectByID(ctx, params.ProjectID, requestPayload)
		handler.HandleResult(w, r, ctx, err, updatedEntity)
	}
}

// @Id				getProject
// @Summary		Get project
// @Description	Get project information by project ID
// @Tags			project
// @Produce		json
// @Param			projectID	path		uint									true	"Project ID"
// @Success		200			{object}	handler.Response{data=entity.Project}	"Success"
// @Failure		400			{object}	error									"Bad Request"
// @Failure		401			{object}	error									"Unauthorized"
// @Failure		429			{object}	error									"Too Many Requests"
// @Failure		404			{object}	error									"Not Found"
// @Failure		500			{object}	error									"Internal Server Error"
// @Router			/api/v1/projects/{projectID} [get]
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

// @Id				listProject
// @Summary		List projects
// @Description	List all or a subset of the projects
// @Tags			project
// @Produce		json
// @Param			orgID		query		uint														false	"OrganizationID to filter project list by. Default to all projects."
// @Param			name		query		string														false	"Project name to filter project list by. This should only return one result if set."
// @Param			fuzzyName	query		string														false	"Fuzzy match project name to filter project list by."
// @Param			page		query		uint														false	"The current page to fetch. Default to 1"
// @Param			pageSize	query		uint														false	"The size of the page. Default to 10"
// @Success		200			{object}	handler.Response{data=[]response.PaginatedProjectResponse}	"Success"
// @Failure		400			{object}	error														"Bad Request"
// @Failure		401			{object}	error														"Unauthorized"
// @Failure		429			{object}	error														"Too Many Requests"
// @Failure		404			{object}	error														"Not Found"
// @Failure		500			{object}	error														"Internal Server Error"
// @Router			/api/v1/projects [get]
func (h *Handler) ListProjects() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Listing project...")

		query := r.URL.Query()
		filter, projectSortOptions, err := h.projectManager.BuildProjectFilterAndSortOptions(ctx, &query)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		projectEntities, err := h.projectManager.ListProjects(ctx, filter, projectSortOptions)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		paginatedResponse := response.PaginatedProjectResponse{
			Projects:    projectEntities.Projects,
			Total:       projectEntities.Total,
			CurrentPage: filter.Pagination.Page,
			PageSize:    filter.Pagination.PageSize,
		}
		handler.HandleResult(w, r, ctx, err, paginatedResponse)
	}
}

func requestHelper(r *http.Request) (context.Context, *httplog.Logger, *ProjectRequestParams, error) {
	ctx := r.Context()
	projectID := chi.URLParam(r, "projectID")
	// Get project with repository
	id, err := strconv.Atoi(projectID)
	if err != nil {
		return nil, nil, nil, constant.ErrInvalidProjectID
	}
	logger := logutil.GetLogger(ctx)
	params := ProjectRequestParams{
		ProjectID: uint(id),
	}
	return ctx, logger, &params, nil
}

package project

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

		// Convert request payload to domain model
		var createdEntity entity.Project
		if err := copier.Copy(&createdEntity, &requestPayload); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		createdEntity.CreationTimestamp = time.Now()
		createdEntity.UpdateTimestamp = time.Now()

		// Get source by id
		sourceEntity, err := h.sourceRepo.Get(ctx, requestPayload.SourceID)
		if err != nil && err == gorm.ErrRecordNotFound {
			render.Render(w, r, handler.FailureResponse(ctx, ErrSourceNotFound))
			return
		} else if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		createdEntity.Source = sourceEntity

		// Get org by id
		organizationEntity, err := h.organizationRepo.Get(ctx, requestPayload.OrganizationID)
		if err != nil && err == gorm.ErrRecordNotFound {
			render.Render(w, r, handler.FailureResponse(ctx, ErrOrgNotFound))
			return
		} else if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		createdEntity.Organization = organizationEntity

		// Create project with repository
		err = h.projectRepo.Create(ctx, &createdEntity)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
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
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Deleting source...")
		projectID := chi.URLParam(r, "projectID")

		// Delete project with repository
		id, err := strconv.Atoi(projectID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidProjectID))
			return
		}
		err = h.projectRepo.Delete(ctx, uint(id))
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
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
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Updating project...")
		projectID := chi.URLParam(r, "projectID")

		// convert project ID to int
		id, err := strconv.Atoi(projectID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidProjectID))
			return
		}

		// Decode the request body into the payload.
		var requestPayload request.UpdateProjectRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		// fmt.Printf("requestPayload.SourceID: %v; requestPayload.Organization: %v", requestPayload.SourceID, requestPayload.OrganizationID)

		// Convert request payload to domain model
		var requestEntity entity.Project
		if err := copier.Copy(&requestEntity, &requestPayload); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Get source by id
		sourceEntity, err := handler.GetSourceByID(ctx, h.sourceRepo, requestPayload.SourceID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		requestEntity.Source = sourceEntity

		// Get organization by id
		organizationEntity, err := handler.GetOrganizationByID(ctx, h.organizationRepo, requestPayload.OrganizationID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		requestEntity.Organization = organizationEntity

		// Get the existing project by id
		updatedEntity, err := h.projectRepo.Get(ctx, uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrUpdatingNonExistingProject))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Overwrite non-zero values in request entity to existing entity
		copier.CopyWithOption(updatedEntity, requestEntity, copier.Option{IgnoreEmpty: true})
		// fmt.Printf("updatedEntity.Source: %v; updatedEntity.Organization: %v", updatedEntity.Source, updatedEntity.Organization)

		// Update project with repository
		err = h.projectRepo.Update(ctx, updatedEntity)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return updated project
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
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Getting project...")
		projectID := chi.URLParam(r, "projectID")

		// Get project with repository
		id, err := strconv.Atoi(projectID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidProjectID))
			return
		}
		existingEntity, err := h.projectRepo.Get(ctx, uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrGettingNonExistingProject))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return found project
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

		projectEntities, err := h.projectRepo.List(ctx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrGettingNonExistingProject))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return found projects
		handler.HandleResult(w, r, ctx, err, projectEntities)
	}
}

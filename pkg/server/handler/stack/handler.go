package stack

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/server/handler"
	"kusionstack.io/kusion/pkg/server/util"
)

// @Summary      Create stack
// @Description  Create a new stack
// @Accept       json
// @Produce      json
// @Param        stack  body      CreateStackRequest  true  "Created stack"
// @Success      200        {object}  entity.Stack        "Success"
// @Failure      400        {object}  errors.DetailError      "Bad Request"
// @Failure      401        {object}  errors.DetailError      "Unauthorized"
// @Failure      429        {object}  errors.DetailError      "Too Many Requests"
// @Failure      404        {object}  errors.DetailError      "Not Found"
// @Failure      500        {object}  errors.DetailError      "Internal Server Error"
// @Router       /api/v1/project/{projectName}/stack/{stackName} [post]
func (h *Handler) CreateStack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Creating stack...")
		// workspaceParam := chi.URLParam(r, "workspaceName")

		// Decode the request body into the payload.
		var requestPayload request.CreateStackRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Convert request payload to domain model
		var createdEntity entity.Stack
		if err := copier.Copy(&createdEntity, &requestPayload); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		// The default state is UnSynced
		createdEntity.SyncState = constant.StackStateUnSynced
		createdEntity.CreationTimestamp = time.Now()
		createdEntity.UpdateTimestamp = time.Now()
		createdEntity.LastSyncTimestamp = time.Unix(0, 0) // default to none

		// TODO: Only project ID should be needed here. Not source and org IDs.
		// Get source by id
		// sourceEntity, err := handler.GetSourceByID(ctx, h.sourceRepo, requestPayload.SourceID)
		// if err != nil {
		// 	render.Render(w, r, handler.FailureResponse(ctx, err))
		// 	return
		// }
		// createdEntity.Source = sourceEntity

		// Get project by id
		projectEntity, err := handler.GetProjectByID(ctx, h.projectRepo, requestPayload.ProjectID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		createdEntity.Project = projectEntity

		// // Get organization by id
		// organizationEntity, err := handler.GetOrganizationByID(ctx, h.orgRepository, requestPayload.OrganizationID)
		// if err != nil {
		// 	render.Render(w, r, handler.FailureResponse(ctx, err))
		// 	return
		// }
		// createdEntity.Organization = organizationEntity
		// TODO: Only project ID should be needed here. Not source and org IDs.

		// Create stack with repository
		err = h.stackRepo.Create(ctx, &createdEntity)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		handler.HandleResult(w, r, ctx, err, createdEntity)
	}
}

// @Summary      Delete stack
// @Description  Delete specified stack by ID
// @Produce      json
// @Param        id   path      int                 true  "Stack ID"
// @Success      200  {object}  entity.Stack       "Success"
// @Failure      400             {object}  errors.DetailError   "Bad Request"
// @Failure      401             {object}  errors.DetailError   "Unauthorized"
// @Failure      429             {object}  errors.DetailError   "Too Many Requests"
// @Failure      404             {object}  errors.DetailError   "Not Found"
// @Failure      500             {object}  errors.DetailError   "Internal Server Error"
// @Router       /api/v1/project/{projectName}/stack/{stackName}  [delete]
// @Router       /api/v1/stack/{stackID} [delete]
func (h *Handler) DeleteStack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Deleting source...")
		stackID := chi.URLParam(r, "stackID")

		// Delete stack with repository
		id, err := strconv.Atoi(stackID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidStacktID))
			return
		}
		err = h.stackRepo.Delete(ctx, uint(id))
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		handler.HandleResult(w, r, ctx, err, "Deletion Success")
	}
}

// @Summary      Update stack
// @Description  Update the specified stack
// @Accept       json
// @Produce      json
// @Param        stack  body      UpdateStackRequest  true  "Updated stack"
// @Success      200     {object}  entity.Stack        "Success"
// @Failure      400     {object}  errors.DetailError   "Bad Request"
// @Failure      401     {object}  errors.DetailError   "Unauthorized"
// @Failure      429     {object}  errors.DetailError   "Too Many Requests"
// @Failure      404     {object}  errors.DetailError   "Not Found"
// @Failure      500     {object}  errors.DetailError   "Internal Server Error"
// @Router       /api/v1/stack/{stackID} [put]
func (h *Handler) UpdateStack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Updating stack...")
		stackID := chi.URLParam(r, "stackID")

		// convert stack ID to int
		id, err := strconv.Atoi(stackID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidStacktID))
			return
		}

		// Decode the request body into the payload.
		var requestPayload request.UpdateStackRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Convert request payload to domain model
		var requestEntity entity.Stack
		if err := copier.Copy(&requestEntity, &requestPayload); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// TODO: Only project ID should be needed here. Not source and org IDs.
		// Get source by id
		// sourceEntity, err := handler.GetSourceByID(ctx, h.sourceRepo, requestPayload.SourceID)
		// if err != nil {
		// 	render.Render(w, r, handler.FailureResponse(ctx, err))
		// 	return
		// }
		// requestEntity.Source = sourceEntity

		// Get project by id
		projectEntity, err := handler.GetProjectByID(ctx, h.projectRepo, requestPayload.ProjectID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		requestEntity.Project = projectEntity

		// // Get organization by id
		// organizationEntity, err := handler.GetOrganizationByID(ctx, h.orgRepository, requestPayload.OrganizationID)
		// if err != nil {
		// 	render.Render(w, r, handler.FailureResponse(ctx, err))
		// 	return
		// }
		// requestEntity.Organization = organizationEntity
		// TODO: Only project ID should be needed here. Not source and org IDs.

		// Get the existing stack by id
		updatedEntity, err := h.stackRepo.Get(ctx, uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrUpdatingNonExistingStack))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Overwrite non-zero values in request entity to existing entity
		copier.CopyWithOption(updatedEntity, requestEntity, copier.Option{IgnoreEmpty: true})

		// Update stack with repository
		err = h.stackRepo.Update(ctx, updatedEntity)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return updated stack
		handler.HandleResult(w, r, ctx, err, updatedEntity)
	}
}

// @Summary      Get stack
// @Description  Get stack information by stack ID
// @Produce      json
// @Param        id   path      int                 true  "Stack ID"
// @Success      200  {object}  entity.Stack       "Success"
// @Failure      400  {object}  errors.DetailError  "Bad Request"
// @Failure      401  {object}  errors.DetailError  "Unauthorized"
// @Failure      429  {object}  errors.DetailError  "Too Many Requests"
// @Failure      404  {object}  errors.DetailError  "Not Found"
// @Failure      500  {object}  errors.DetailError  "Internal Server Error"
// @Router       /api/v1/stack/{stackID} [get]
func (h *Handler) GetStack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Getting stack...")
		stackID := chi.URLParam(r, "stackID")

		// Get stack with repository
		id, err := strconv.Atoi(stackID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidStacktID))
			return
		}
		existingEntity, err := h.stackRepo.Get(ctx, uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrGettingNonExistingStack))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return found stack
		handler.HandleResult(w, r, ctx, err, existingEntity)
	}
}

// @Summary      List stacks
// @Description  List all stacks
// @Produce      json
// @Success      200  {object}  entity.Stack       "Success"
// @Failure      400  {object}  errors.DetailError  "Bad Request"
// @Failure      401  {object}  errors.DetailError  "Unauthorized"
// @Failure      429  {object}  errors.DetailError  "Too Many Requests"
// @Failure      404  {object}  errors.DetailError  "Not Found"
// @Failure      500  {object}  errors.DetailError  "Internal Server Error"
// @Router       /api/v1/stack [get]
func (h *Handler) ListStacks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Listing stack...")

		stackEntities, err := h.stackRepo.List(ctx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrGettingNonExistingStack))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return found stacks
		handler.HandleResult(w, r, ctx, err, stackEntities)
	}
}

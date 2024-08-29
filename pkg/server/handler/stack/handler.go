package stack

import (
	"net/http"

	"github.com/go-chi/render"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/server/handler"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

//	@Id				createStack
//	@Summary		Create stack
//	@Description	Create a new stack
//	@Tags			stack
//	@Accept			json
//	@Produce		json
//	@Param			stack			body		request.CreateStackRequest	true	"Created stack"
//	@Param			fromTemplate	query		bool						false	"Whether to create an AppConfig from template when creating the stack"
//	@Param			initTopology	query		bool						false	"Whether to initialize an AppTopology from template when creating the stack"
//	@Success		200				{object}	entity.Stack				"Success"
//	@Failure		400				{object}	error						"Bad Request"
//	@Failure		401				{object}	error						"Unauthorized"
//	@Failure		429				{object}	error						"Too Many Requests"
//	@Failure		404				{object}	error						"Not Found"
//	@Failure		500				{object}	error						"Internal Server Error"
//	@Router			/api/v1/stacks [post]
func (h *Handler) CreateStack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Creating stack...")

		// Decode the request body into the payload.
		var requestPayload request.CreateStackRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Validate request payload
		if err := requestPayload.Validate(); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Create entity
		createdEntity, err := h.stackManager.CreateStack(ctx, requestPayload)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		defer func() {
			if err != nil {
				// Rollback
				err = h.stackManager.DeleteStackByID(ctx, createdEntity.ID)
				if err != nil {
					logger.Info("Failed to rollback stack creation", "stackID", createdEntity.ID, "error", err)
				}
			}
		}()

		handler.HandleResult(w, r, ctx, err, createdEntity)
	}
}

//	@Id				deleteStack
//	@Summary		Delete stack
//	@Description	Delete specified stack by ID
//	@Tags			stack
//	@Produce		json
//	@Param			stack_id	path		int		true	"Stack ID"
//	@Success		200			{object}	string	"Success"
//	@Failure		400			{object}	error	"Bad Request"
//	@Failure		401			{object}	error	"Unauthorized"
//	@Failure		429			{object}	error	"Too Many Requests"
//	@Failure		404			{object}	error	"Not Found"
//	@Failure		500			{object}	error	"Internal Server Error"
//	@Router			/api/v1/stacks/{stack_id} [delete]
func (h *Handler) DeleteStack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Deleting source...", "stackID", params.StackID)

		err = h.stackManager.DeleteStackByID(ctx, params.StackID)
		handler.HandleResult(w, r, ctx, err, "Deletion Success")
	}
}

//	@Id				updateStack
//	@Summary		Update stack
//	@Description	Update the specified stack
//	@Tags			stack
//	@Accept			json
//	@Produce		json
//	@Param			stack_id	path		int							true	"Stack ID"
//	@Param			stack		body		request.UpdateStackRequest	true	"Updated stack"
//	@Success		200			{object}	entity.Stack				"Success"
//	@Failure		400			{object}	error						"Bad Request"
//	@Failure		401			{object}	error						"Unauthorized"
//	@Failure		429			{object}	error						"Too Many Requests"
//	@Failure		404			{object}	error						"Not Found"
//	@Failure		500			{object}	error						"Internal Server Error"
//	@Router			/api/v1/stacks/{stack_id} [put]
func (h *Handler) UpdateStack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Updating stack...", "stackID", params.StackID)

		// Decode the request body into the payload.
		var requestPayload request.UpdateStackRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		updatedEntity, err := h.stackManager.UpdateStackByID(ctx, params.StackID, requestPayload)
		handler.HandleResult(w, r, ctx, err, updatedEntity)
	}
}

//	@Id				getStack
//	@Summary		Get stack
//	@Description	Get stack information by stack ID
//	@Tags			stack
//	@Produce		json
//	@Param			stack_id	path		int				true	"Stack ID"
//	@Success		200			{object}	entity.Stack	"Success"
//	@Failure		400			{object}	error			"Bad Request"
//	@Failure		401			{object}	error			"Unauthorized"
//	@Failure		429			{object}	error			"Too Many Requests"
//	@Failure		404			{object}	error			"Not Found"
//	@Failure		500			{object}	error			"Internal Server Error"
//	@Router			/api/v1/stacks/{stack_id} [get]
func (h *Handler) GetStack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Getting stack...", "stackID", params.StackID)

		existingEntity, err := h.stackManager.GetStackByID(ctx, params.StackID)
		handler.HandleResult(w, r, ctx, err, existingEntity)
	}
}

//	@Id				listStack
//	@Summary		List stacks
//	@Description	List all stacks
//	@Tags			stack
//	@Produce		json
//	@Param			projectID			query		uint			false	"ProjectID to filter stacks by. Default to all"
//	@Param			orgID				query		uint			false	"OrgID to filter stacks by. Default to all"
//	@Param			projectName			query		string			false	"ProjectName to filter stacks by. Default to all"
//	@Param			cloud				query		string			false	"Cloud to filter stacks by. Default to all"
//	@Param			env					query		string			false	"Environment to filter stacks by. Default to all"
//	@Param			getLastSyncedBase	query		bool			false	"Whether to get last synced base revision. Default to false"
//	@Success		200					{object}	[]entity.Stack	"Success"
//	@Failure		400					{object}	error			"Bad Request"
//	@Failure		401					{object}	error			"Unauthorized"
//	@Failure		429					{object}	error			"Too Many Requests"
//	@Failure		404					{object}	error			"Not Found"
//	@Failure		500					{object}	error			"Internal Server Error"
//	@Router			/api/v1/stacks [get]
func (h *Handler) ListStacks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Listing stack...")

		orgIDParam := r.URL.Query().Get("orgID")
		projectIDParam := r.URL.Query().Get("projectID")
		projectNameParam := r.URL.Query().Get("projectName")
		envParam := r.URL.Query().Get("env")

		filter, err := h.stackManager.BuildStackFilter(ctx, orgIDParam, projectIDParam, projectNameParam, envParam)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		stackEntities, err := h.stackManager.ListStacks(ctx, filter)
		handler.HandleResult(w, r, ctx, err, stackEntities)
	}
}

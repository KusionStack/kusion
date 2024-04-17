package stack

import (
	"net/http"

	"github.com/go-chi/render"
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

		// Decode the request body into the payload.
		var requestPayload request.CreateStackRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		createdEntity, err := h.stackManager.CreateStack(ctx, requestPayload)
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

		stackEntities, err := h.stackManager.ListStacks(ctx)
		handler.HandleResult(w, r, ctx, err, stackEntities)
	}
}

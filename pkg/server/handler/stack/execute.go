package stack

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-logr/logr"
	yamlv2 "gopkg.in/yaml.v2"
	"kusionstack.io/kusion/pkg/server/handler"
	stackmanager "kusionstack.io/kusion/pkg/server/manager/stack"
	"kusionstack.io/kusion/pkg/server/util"
)

// @Summary      Preview stack
// @Description  Preview stack information by stack ID
// @Produce      json
// @Param        id   path      int                 true  "Stack ID"
// @Success      200  {object}  entity.Stack       "Success"
// @Failure      400  {object}  errors.DetailError  "Bad Request"
// @Failure      401  {object}  errors.DetailError  "Unauthorized"
// @Failure      429  {object}  errors.DetailError  "Too Many Requests"
// @Failure      404  {object}  errors.DetailError  "Not Found"
// @Failure      500  {object}  errors.DetailError  "Internal Server Error"
// @Router       /api/v1/stack/{stackID}/preview [post]
func (h *Handler) PreviewStack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Previewing stack...", "stackID", params.StackID)

		// Call preview stack
		changes, err := h.stackManager.PreviewStack(ctx, params.StackID, params.Workspace)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		previewChanges, err := stackmanager.ProcessChanges(ctx, w, changes, params.Format, params.Detail)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		render.Render(w, r, handler.SuccessResponse(ctx, previewChanges))
	}
}

// @Summary      Generate stack
// @Description  Generate stack information by stack ID
// @Produce      json
// @Param        id   path      int                 true  "Stack ID"
// @Success      200  {object}  entity.Stack       "Success"
// @Failure      400  {object}  errors.DetailError  "Bad Request"
// @Failure      401  {object}  errors.DetailError  "Unauthorized"
// @Failure      429  {object}  errors.DetailError  "Too Many Requests"
// @Failure      404  {object}  errors.DetailError  "Not Found"
// @Failure      500  {object}  errors.DetailError  "Internal Server Error"
// @Router       /api/v1/stack/{stackID}/generate [post]
func (h *Handler) GenerateStack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Generating stack...", "stackID", params.StackID)

		// Call generate stack
		sp, err := h.stackManager.GenerateStack(ctx, params.StackID, params.Workspace)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		yaml, err := yamlv2.Marshal(sp)
		handler.HandleResult(w, r, ctx, err, string(yaml))
	}
}

// @Summary      Apply stack
// @Description  Apply stack information by stack ID
// @Produce      json
// @Param        id   path      int                 true  "Stack ID"
// @Success      200  {object}  entity.Stack       "Success"
// @Failure      400  {object}  errors.DetailError  "Bad Request"
// @Failure      401  {object}  errors.DetailError  "Unauthorized"
// @Failure      429  {object}  errors.DetailError  "Too Many Requests"
// @Failure      404  {object}  errors.DetailError  "Not Found"
// @Failure      500  {object}  errors.DetailError  "Internal Server Error"
// @Router       /api/v1/stack/{stackID}/apply [post]
func (h *Handler) ApplyStack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Applying stack...", "stackID", params.StackID)

		err = h.stackManager.ApplyStack(ctx, params.StackID, params.Workspace, params.Format, params.Detail, params.Dryrun, w)
		if err != nil {
			if err == stackmanager.ErrDryrunDestroy {
				render.Render(w, r, handler.SuccessResponse(ctx, "Dry-run mode enabled, the above resources will be destroyed if dryrun is set to false"))
				return
			} else {
				render.Render(w, r, handler.FailureResponse(ctx, err))
				return
			}
		}

		// Apply completed
		logger.Info("apply completed")
		render.Render(w, r, handler.SuccessResponse(ctx, "apply completed"))

		// TODO: How to implement watch?
		// if o.Watch {
		// 	fmt.Println("Start watching changes ...")
		// 	if err = Watch(o, sp, changes); err != nil {
		// 		return err
		// 	}
		// }
	}
}

// @Summary      Destroy stack
// @Description  Destroy stack information by stack ID
// @Produce      json
// @Param        id   path      int                 true  "Stack ID"
// @Success      200  {object}  entity.Stack       "Success"
// @Failure      400  {object}  errors.DetailError  "Bad Request"
// @Failure      401  {object}  errors.DetailError  "Unauthorized"
// @Failure      429  {object}  errors.DetailError  "Too Many Requests"
// @Failure      404  {object}  errors.DetailError  "Not Found"
// @Failure      500  {object}  errors.DetailError  "Internal Server Error"
// @Router       /api/v1/stack/{stackID}/destroy [post]
func (h *Handler) DestroyStack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Destroying stack...", "stackID", params.StackID)

		err = h.stackManager.DestroyStack(ctx, params.StackID, params.Workspace, params.Detail, params.Dryrun, w)
		if err != nil {
			if err == stackmanager.ErrDryrunDestroy {
				render.Render(w, r, handler.SuccessResponse(ctx, "Dry-run mode enabled, the above resources will be destroyed if dryrun is set to false"))
				return
			} else {
				render.Render(w, r, handler.FailureResponse(ctx, err))
				return
			}
		}

		// Destroy completed
		logger.Info("destroy completed")
		render.Render(w, r, handler.SuccessResponse(ctx, "destroy completed"))
	}
}

func requestHelper(r *http.Request) (context.Context, *logr.Logger, *StackRequestParams, error) {
	ctx := r.Context()
	stackID := chi.URLParam(r, "stackID")
	// Get stack with repository
	id, err := strconv.Atoi(stackID)
	if err != nil {
		return nil, nil, nil, stackmanager.ErrInvalidStackID
	}
	logger := util.GetLogger(ctx)
	// Get Params
	detailParam, _ := strconv.ParseBool(r.URL.Query().Get("detail"))
	dryrunParam, _ := strconv.ParseBool(r.URL.Query().Get("dryrun"))
	outputParam := r.URL.Query().Get("output")
	// TODO: Should match automatically eventually???
	workspaceParam := r.URL.Query().Get("workspace")
	params := StackRequestParams{
		StackID:   uint(id),
		Workspace: workspaceParam,
		Detail:    detailParam,
		Dryrun:    dryrunParam,
		Format:    outputParam,
	}
	return ctx, &logger, &params, nil
}

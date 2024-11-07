package stack

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/render"

	yamlv2 "gopkg.in/yaml.v2"

	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/server/handler"
	stackmanager "kusionstack.io/kusion/pkg/server/manager/stack"
	appmiddleware "kusionstack.io/kusion/pkg/server/middleware"
	authutil "kusionstack.io/kusion/pkg/server/util/auth"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

// @Id				previewStack
// @Summary		Preview stack
// @Description	Preview stack information by stack ID
// @Tags			stack
// @Produce		json
// @Param			stack_id	path		int				true	"Stack ID"
// @Param			output		query		string			false	"Output format. Choices are: json, default. Default to default output format in Kusion."
// @Param			detail		query		bool			false	"Show detailed output"
// @Param			specID		query		string			false	"The Spec ID to use for the preview. Default to the last one generated."
// @Param			force		query		bool			false	"Force the preview even when the stack is locked"
// @Success		200			{object}	models.Changes	"Success"
// @Failure		400			{object}	error			"Bad Request"
// @Failure		401			{object}	error			"Unauthorized"
// @Failure		429			{object}	error			"Too Many Requests"
// @Failure		404			{object}	error			"Not Found"
// @Failure		500			{object}	error			"Internal Server Error"
// @Router			/api/v1/stacks/{stack_id}/preview [post]
func (h *Handler) PreviewStack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Previewing stack...", "stackID", params.StackID)

		var requestPayload request.StackImportRequest
		if params.ExecuteParams.ImportResources {
			if err := requestPayload.Decode(r); err != nil {
				if err == io.EOF {
					render.Render(w, r, handler.FailureResponse(ctx, fmt.Errorf("request body should not be empty when importResources is set to true")))
					return
				} else {
					render.Render(w, r, handler.FailureResponse(ctx, err))
					return
				}
			}
		}

		// Call preview stack
		changes, err := h.stackManager.PreviewStack(ctx, params, requestPayload)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		previewChanges, err := stackmanager.ProcessChanges(ctx, w, changes, params.Format, params.ExecuteParams.Detail)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		render.Render(w, r, handler.SuccessResponse(ctx, previewChanges))
	}
}

func CopyToNewContext(ctx context.Context) context.Context {
	newCtx := context.Background()
	newCtx = context.WithValue(newCtx, appmiddleware.TraceIDKey, appmiddleware.GetTraceID(ctx))
	newCtx = context.WithValue(newCtx, appmiddleware.UserIDKey, appmiddleware.GetUserID(ctx))
	if logger, ok := ctx.Value(appmiddleware.APILoggerKey).(*httplog.Logger); ok {
		newCtx = context.WithValue(newCtx, appmiddleware.APILoggerKey, logger)
	}
	return newCtx
}

// @Id				generateStack
// @Summary		Generate stack
// @Description	Generate stack information by stack ID
// @Tags			stack
// @Produce		json
// @Param			stack_id	path		int		true	"Stack ID"
// @Param			format		query		string	false	"The format to generate the spec in. Choices are: spec. Default to spec."
// @Param			force		query		bool	false	"Force the generate even when the stack is locked"
// @Success		200			{object}	v1.Spec	"Success"
// @Failure		400			{object}	error	"Bad Request"
// @Failure		401			{object}	error	"Unauthorized"
// @Failure		429			{object}	error	"Too Many Requests"
// @Failure		404			{object}	error	"Not Found"
// @Failure		500			{object}	error	"Internal Server Error"
// @Router			/api/v1/stacks/{stack_id}/generate [post]
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
		_, sp, err := h.stackManager.GenerateSpec(ctx, params)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		yaml, err := yamlv2.Marshal(sp)
		handler.HandleResult(w, r, ctx, err, string(yaml))
	}
}

// @Id				applyStack
// @Summary		Apply stack
// @Description	Apply stack information by stack ID
// @Tags			stack
// @Produce		json
// @Param			stack_id	path		int		true	"Stack ID"
// @Param			specID		query		string	false	"The Spec ID to use for the apply. Will generate a new spec if omitted."
// @Param			force		query		bool	false	"Force the apply even when the stack is locked. May cause concurrency issues!!!"
// @Param			dryrun		query		bool	false	"Apply in dry-run mode"
// @Success		200			{object}	string	"Success"
// @Failure		400			{object}	error	"Bad Request"
// @Failure		401			{object}	error	"Unauthorized"
// @Failure		429			{object}	error	"Too Many Requests"
// @Failure		404			{object}	error	"Not Found"
// @Failure		500			{object}	error	"Internal Server Error"
// @Router			/api/v1/stacks/{stack_id}/apply [post]
func (h *Handler) ApplyStack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Applying stack...", "stackID", params.StackID)

		var requestPayload request.StackImportRequest
		if params.ExecuteParams.ImportResources {
			if err := requestPayload.Decode(r); err != nil {
				if err == io.EOF {
					render.Render(w, r, handler.FailureResponse(ctx, fmt.Errorf("request body should not be empty when importResources is set to true")))
					return
				} else {
					render.Render(w, r, handler.FailureResponse(ctx, err))
					return
				}
			}
		}

		err = h.stackManager.ApplyStack(ctx, params, requestPayload)
		if err != nil {
			if err == stackmanager.ErrDryrunDestroy {
				render.Render(w, r, handler.SuccessResponse(ctx, "Dry-run mode enabled, the above resources will be applied if dryrun is set to false"))
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

// @Id				destroyStack
// @Summary		Destroy stack
// @Description	Destroy stack information by stack ID
// @Tags			stack
// @Produce		json
// @Param			stack_id	path		int		true	"Stack ID"
// @Param			force		query		bool	false	"Force the destroy even when the stack is locked. May cause concurrency issues!!!"
// @Param			dryrun		query		bool	false	"Destroy in dry-run mode"
// @Success		200			{object}	string	"Success"
// @Failure		400			{object}	error	"Bad Request"
// @Failure		401			{object}	error	"Unauthorized"
// @Failure		429			{object}	error	"Too Many Requests"
// @Failure		404			{object}	error	"Not Found"
// @Failure		500			{object}	error	"Internal Server Error"
// @Router			/api/v1/stacks/{stack_id}/destroy [post]
func (h *Handler) DestroyStack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Destroying stack...", "stackID", params.StackID)

		err = h.stackManager.DestroyStack(ctx, params, w)
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

func requestHelper(r *http.Request) (context.Context, *httplog.Logger, *stackmanager.StackRequestParams, error) {
	ctx := r.Context()
	stackID := chi.URLParam(r, "stackID")
	// Get stack with repository
	id, err := strconv.Atoi(stackID)
	if err != nil {
		return nil, nil, nil, stackmanager.ErrInvalidStackID
	}
	logger := logutil.GetLogger(ctx)
	// Get Params
	outputParam := r.URL.Query().Get("output")
	detailParam, _ := strconv.ParseBool(r.URL.Query().Get("detail"))
	dryrunParam, _ := strconv.ParseBool(r.URL.Query().Get("dryrun"))
	forceParam, _ := strconv.ParseBool(r.URL.Query().Get("force"))
	importResourcesParam, _ := strconv.ParseBool(r.URL.Query().Get("importResources"))
	specIDParam := r.URL.Query().Get("specID")
	// TODO: Should match automatically eventually???
	workspaceParam := r.URL.Query().Get("workspace")
	operatorParam, err := authutil.GetSubjectFromUnverifiedJWTToken(ctx, r)
	// fall back to x-kusion-user if operator is not parsed from cookie
	if operatorParam == "" || err != nil {
		operatorParam = appmiddleware.GetUserID(ctx)
		if operatorParam == "" {
			operatorParam = constant.DefaultUser
		}
	}
	if workspaceParam == "" {
		workspaceParam = constant.DefaultWorkspace
	}
	executeParams := stackmanager.StackExecuteParams{
		Detail:          detailParam,
		Dryrun:          dryrunParam,
		Force:           forceParam,
		SpecID:          specIDParam,
		ImportResources: importResourcesParam,
	}
	params := stackmanager.StackRequestParams{
		StackID:       uint(id),
		Workspace:     workspaceParam,
		Format:        outputParam,
		Operator:      operatorParam,
		ExecuteParams: executeParams,
	}
	return ctx, logger, &params, nil
}

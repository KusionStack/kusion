package stack

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/render"

	yamlv2 "gopkg.in/yaml.v2"
	_ "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"

	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/server/handler"
	stackmanager "kusionstack.io/kusion/pkg/server/manager/stack"
)

// @Id				previewStack
// @Summary		Preview stack
// @Description	Preview stack information by stack ID
// @Tags			stack
// @Produce		json
// @Param			stackID				path		int										true	"Stack ID"
// @Param			importedResources	body		request.StackImportRequest				false	"The resources to import during the stack preview"
// @Param			workspace			query		string									true	"The target workspace to preview the spec in."
// @Param			importResources		query		bool									false	"Import existing resources during the stack preview"
// @Param			output				query		string									false	"Output format. Choices are: json, default. Default to default output format in Kusion."
// @Param			detail				query		bool									false	"Show detailed output"
// @Param			specID				query		string									false	"The Spec ID to use for the preview. Default to the last one generated."
// @Param			force				query		bool									false	"Force the preview even when the stack is locked"
// @Success		200					{object}	handler.Response{data=models.Changes}	"Success"
// @Failure		400					{object}	error									"Bad Request"
// @Failure		401					{object}	error									"Unauthorized"
// @Failure		429					{object}	error									"Too Many Requests"
// @Failure		404					{object}	error									"Not Found"
// @Failure		500					{object}	error									"Internal Server Error"
// @Router			/api/v1/stacks/{stackID}/preview [post]
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

// @Id				generateStack
// @Summary		Generate stack
// @Description	Generate stack information by stack ID
// @Tags			stack
// @Produce		json
// @Param			stackID		path		int								true	"Stack ID"
// @Param			workspace	query		string							true	"The target workspace to preview the spec in."
// @Param			format		query		string							false	"The format to generate the spec in. Choices are: spec. Default to spec."
// @Param			force		query		bool							false	"Force the generate even when the stack is locked"
// @Success		200			{object}	handler.Response{data=v1.Spec}	"Success"
// @Failure		400			{object}	error							"Bad Request"
// @Failure		401			{object}	error							"Unauthorized"
// @Failure		429			{object}	error							"Too Many Requests"
// @Failure		404			{object}	error							"Not Found"
// @Failure		500			{object}	error							"Internal Server Error"
// @Router			/api/v1/stacks/{stackID}/generate [post]
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
// @Param			stackID				path		int								true	"Stack ID"
// @Param			importedResources	body		request.StackImportRequest		false	"The resources to import during the stack preview"
// @Param			workspace			query		string							true	"The target workspace to preview the spec in."
// @Param			importResources		query		bool							false	"Import existing resources during the stack preview"
// @Param			specID				query		string							false	"The Spec ID to use for the apply. Will generate a new spec if omitted."
// @Param			force				query		bool							false	"Force the apply even when the stack is locked. May cause concurrency issues!!!"
// @Param			dryrun				query		bool							false	"Apply in dry-run mode"
// @Success		200					{object}	handler.Response{data=string}	"Success"
// @Failure		400					{object}	error							"Bad Request"
// @Failure		401					{object}	error							"Unauthorized"
// @Failure		429					{object}	error							"Too Many Requests"
// @Failure		404					{object}	error							"Not Found"
// @Failure		500					{object}	error							"Internal Server Error"
// @Router			/api/v1/stacks/{stackID}/apply [post]
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
	}
}

// @Id				destroyStack
// @Summary		Destroy stack
// @Description	Destroy stack information by stack ID
// @Tags			stack
// @Produce		json
// @Param			stackID		path		int								true	"Stack ID"
// @Param			workspace	query		string							true	"The target workspace to preview the spec in."
// @Param			force		query		bool							false	"Force the destroy even when the stack is locked. May cause concurrency issues!!!"
// @Param			dryrun		query		bool							false	"Destroy in dry-run mode"
// @Success		200			{object}	handler.Response{data=string}	"Success"
// @Failure		400			{object}	error							"Bad Request"
// @Failure		401			{object}	error							"Unauthorized"
// @Failure		429			{object}	error							"Too Many Requests"
// @Failure		404			{object}	error							"Not Found"
// @Failure		500			{object}	error							"Internal Server Error"
// @Router			/api/v1/stacks/{stackID}/destroy [post]
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

package stack

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/render"

	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/server/handler"
	stackmanager "kusionstack.io/kusion/pkg/server/manager/stack"

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
func (h *Handler) PreviewStackAsync() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Previewing stack...", "stackID", params.StackID)

		// var requestPayload request.StackImportRequest
		var requestPayload request.CreateRunRequest
		if err := requestPayload.Decode(r); err != nil {
			if err == io.EOF {
				render.Render(w, r, handler.FailureResponse(ctx, fmt.Errorf("request body should not be empty when importResources is set to true")))
				return
			} else {
				render.Render(w, r, handler.FailureResponse(ctx, err))
				return
			}
		}

		// Create a Run object in database and start background task
		runEntity, err := h.stackManager.CreateRun(ctx, requestPayload)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		render.Render(w, r, handler.SuccessResponse(ctx, runEntity))

		runLogger := logutil.GetRunLogger(ctx)
		runLoggerBuffer := logutil.GetRunLoggerBuffer(ctx)
		runLogger.Info("Starting previewing stack in StackManager ... This is a preview run.", "runID", runEntity.ID)

		// Starts a safe goroutine using given recover handler
		go func() {
			// defer safe.HandleCrash(aciLoggingRecoverHandler(h.aciClient, &req, log))
			defer func() {
				// update status of the run
				logger.Info("preview completed for stack", "stackID", params.StackID, "time", time.Now())
			}()

			logger.Info("Async preview in progress")
			newCtx := CopyToNewContext(ctx)
			// Call preview stack
			changes, err := h.stackManager.PreviewStack(newCtx, params, requestPayload.ImportedResources)
			if err != nil {
				render.Render(w, r, handler.FailureResponse(ctx, err))
				logger.Error("Error previewing stack", "error", err)
				return
			}

			previewChanges, err := stackmanager.ProcessChanges(ctx, w, changes, params.Format, params.ExecuteParams.Detail)
			if err != nil {
				render.Render(w, r, handler.FailureResponse(ctx, err))
				return
			}

			// time.Sleep(5 * time.Second)

			if pc, ok := previewChanges.(*models.Changes); ok {
				pcBytes, _ := json.Marshal(pc)
				logger.Info("Preview changes", "changes", string(pcBytes))
				fmt.Println(string(pcBytes))
				// Update the Run object in database to include the preview result
				updateRunResultPayload := request.UpdateRunResultRequest{
					Result: string(pcBytes),
					Status: string(constant.RunStatusSucceeded),
					Logs:   runLoggerBuffer.String(),
				}
				_, err := h.stackManager.UpdateRunResultAndStatusByID(newCtx, runEntity.ID, updateRunResultPayload)
				if err != nil {
					logger.Error("Error updating run result", "error", err)
					return
				}
			}
			// render.Render(w, r, handler.SuccessResponse(ctx, previewChanges))
		}()
	}
}

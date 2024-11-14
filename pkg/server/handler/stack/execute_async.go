package stack

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/render"
	yamlv2 "gopkg.in/yaml.v2"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/server/handler"
	stackmanager "kusionstack.io/kusion/pkg/server/manager/stack"

	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

// @Id				previewStackAsync
// @Summary		Asynchronously preview stack
// @Description	Start a run and asynchronously preview stack changes by stack ID
// @Tags			stack
// @Produce		json
// @Param			stack_id			path		int							true	"Stack ID"
// @Param			importedResources	body		request.StackImportRequest	false	"The resources to import during the stack preview"
// @Param			workspace			query		string						true	"The target workspace to preview the spec in."
// @Param			importResources		query		bool						false	"Import existing resources during the stack preview"
// @Param			output				query		string						false	"Output format. Choices are: json, default. Default to default output format in Kusion."
// @Param			detail				query		bool						false	"Show detailed output"
// @Param			specID				query		string						false	"The Spec ID to use for the preview. Default to the last one generated."
// @Param			force				query		bool						false	"Force the preview even when the stack is locked"
// @Success		200					{object}	entity.Run					"Success"
// @Failure		400					{object}	error						"Bad Request"
// @Failure		401					{object}	error						"Unauthorized"
// @Failure		429					{object}	error						"Too Many Requests"
// @Failure		404					{object}	error						"Not Found"
// @Failure		500					{object}	error						"Internal Server Error"
// @Router			/api/v1/stacks/{stack_id}/preview [post]
func (h *Handler) PreviewStackAsync() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Previewing stack asynchronously...", "stackID", params.StackID)

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

		runLogger := logutil.GetRunLogger(ctx)
		runLogger.Info("Starting previewing stack in StackManager ... This is a preview run.", "runID", runEntity.ID)

		// Starts a safe goroutine using given recover handler
		inBufferZone := h.workerPool.Do(func() {
			// defer safe.HandleCrash(aciLoggingRecoverHandler(h.aciClient, &req, log))
			logger.Info("Async preview in progress")
			var previewChanges any
			newCtx, cancel := CopyToNewContextWithTimeout(ctx, constant.RunTimeOut)
			defer cancel() // make sure the context is canceled to free resources

			// update status of the run when exiting the async run
			defer func() {
				select {
				case <-newCtx.Done():
					logger.Info("preview execution timed out", "stackID", params.StackID, "time", time.Now(), "timeout", newCtx.Err())
					h.setRunToCancelled(newCtx, runEntity.ID)
				default:
					if err != nil {
						logger.Info("preview failed for stack", "stackID", params.StackID, "time", time.Now())
						h.setRunToFailed(newCtx, runEntity.ID)
					} else {
						logger.Info("preview completed for stack", "stackID", params.StackID, "time", time.Now())
						if pc, ok := previewChanges.(*models.Changes); ok {
							h.setRunToSuccess(newCtx, runEntity.ID, pc)
						} else {
							logger.Error("Error casting preview changes to models.Changes", "error", "casting error")
							h.setRunToFailed(newCtx, runEntity.ID)
						}
					}
				}
			}()

			// Call preview stack
			changes, err := h.stackManager.PreviewStack(newCtx, params, requestPayload.ImportedResources)
			if err != nil {
				logger.Error("Error previewing stack", "error", err)
				return
			}

			previewChanges, err = stackmanager.ProcessChanges(newCtx, w, changes, params.Format, params.ExecuteParams.Detail)
			if err != nil {
				logger.Error("Error processing preview changes", "error", err)
				return
			}
		})
		defer func() {
			if inBufferZone {
				logger.Info("The task is in the buffer zone, waiting for an available worker")
				h.setRunToQueued(ctx, runEntity.ID)
			}
		}()
		render.Render(w, r, handler.SuccessResponse(ctx, runEntity))
	}
}

// @Id				applyStackAsync
// @Summary		Asynchronously apply stack
// @Description	Start a run and asynchronously apply stack changes by stack ID
// @Tags			stack
// @Produce		json
// @Param			stack_id			path		int							true	"Stack ID"
// @Param			importedResources	body		request.StackImportRequest	false	"The resources to import during the stack preview"
// @Param			workspace			query		string						true	"The target workspace to preview the spec in."
// @Param			importResources		query		bool						false	"Import existing resources during the stack preview"
// @Param			specID				query		string						false	"The Spec ID to use for the apply. Will generate a new spec if omitted."
// @Param			force				query		bool						false	"Force the apply even when the stack is locked. May cause concurrency issues!!!"
// @Param			dryrun				query		bool						false	"Apply in dry-run mode"
// @Success		200					{object}	entity.Run					"Success"
// @Failure		400					{object}	error						"Bad Request"
// @Failure		401					{object}	error						"Unauthorized"
// @Failure		429					{object}	error						"Too Many Requests"
// @Failure		404					{object}	error						"Not Found"
// @Failure		500					{object}	error						"Internal Server Error"
// @Router			/api/v1/stacks/{stack_id}/apply/async [post]
func (h *Handler) ApplyStackAsync() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Applying stack asynchronously...", "stackID", params.StackID)

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

		runLogger := logutil.GetRunLogger(ctx)
		runLogger.Info("Starting applying stack in StackManager ... This is an apply run.", "runID", runEntity.ID)

		// Starts a safe goroutine using given recover handler
		inBufferZone := h.workerPool.Do(func() {
			// defer safe.HandleCrash(aciLoggingRecoverHandler(h.aciClient, &req, log))
			logger.Info("Async apply in progress")
			newCtx, cancel := CopyToNewContextWithTimeout(ctx, constant.RunTimeOut)
			defer cancel() // make sure the context is canceled to free resources

			// update status of the run when exiting the async run
			defer func() {
				select {
				case <-newCtx.Done():
					logger.Info("apply execution timed out", "stackID", params.StackID, "time", time.Now(), "timeout", newCtx.Err())
					h.setRunToCancelled(newCtx, runEntity.ID)
				default:
					if err != nil {
						logger.Info("apply failed for stack", "stackID", params.StackID, "time", time.Now())
						h.setRunToFailed(newCtx, runEntity.ID)
					} else {
						logger.Info("apply completed for stack", "stackID", params.StackID, "time", time.Now())
						h.setRunToSuccess(newCtx, runEntity.ID, "apply completed")
					}
				}
			}()

			// call apply stack
			err = h.stackManager.ApplyStack(newCtx, params, requestPayload.ImportedResources)
			if err != nil {
				if err == stackmanager.ErrDryrunDestroy {
					render.Render(w, r, handler.SuccessResponse(ctx, "Dry-run mode enabled, the above resources will be applied if dryrun is set to false"))
					return
				} else {
					logger.Error("Error applying stack", "error", err)
					return
				}
			}
		})

		defer func() {
			if inBufferZone {
				logger.Info("The task is in the buffer zone, waiting for an available worker")
				h.setRunToQueued(ctx, runEntity.ID)
			}
		}()
		render.Render(w, r, handler.SuccessResponse(ctx, runEntity))
		// TODO: How to implement watch?
		// if o.Watch {
		// 	fmt.Println("Start watching changes ...")
		// 	if err = Watch(o, sp, changes); err != nil {
		// 		return err
		// 	}
		// }
	}
}

// @Id				generateStackAsync
// @Summary		Asynchronously generate stack
// @Description	Start a run and asynchronously generate stack spec by stack ID
// @Tags			stack
// @Produce		json
// @Param			stack_id	path		int		true	"Stack ID"
// @Param			workspace			query		string						true	"The target workspace to preview the spec in."
// @Param			format		query		string	false	"The format to generate the spec in. Choices are: spec. Default to spec."
// @Param			force		query		bool	false	"Force the generate even when the stack is locked"
// @Success		200			{object}	v1.Spec	"Success"
// @Failure		400			{object}	error	"Bad Request"
// @Failure		401			{object}	error	"Unauthorized"
// @Failure		429			{object}	error	"Too Many Requests"
// @Failure		404			{object}	error	"Not Found"
// @Failure		500			{object}	error	"Internal Server Error"
// @Router			/api/v1/stacks/{stack_id}/generate/async [post]
func (h *Handler) GenerateStackAsync() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Generating stack asynchronously...", "stackID", params.StackID)

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

		runLogger := logutil.GetRunLogger(ctx)
		runLogger.Info("Starting generating stack in StackManager ... This is a generate run.", "runID", runEntity.ID)

		// Starts a safe goroutine using given recover handler
		inBufferZone := h.workerPool.Do(func() {
			// defer safe.HandleCrash(aciLoggingRecoverHandler(h.aciClient, &req, log))
			logger.Info("Async generate in progress")
			newCtx, cancel := CopyToNewContextWithTimeout(ctx, constant.RunTimeOut)
			var sp *apiv1.Spec
			defer cancel() // make sure the context is canceled to free resources

			// update status of the run when exiting the async run
			defer func() {
				select {
				case <-newCtx.Done():
					logger.Info("generate execution timed out", "stackID", params.StackID, "time", time.Now(), "timeout", newCtx.Err())
					h.setRunToCancelled(newCtx, runEntity.ID)
				default:
					if err != nil {
						logger.Info("generate failed for stack", "stackID", params.StackID, "time", time.Now())
						h.setRunToFailed(newCtx, runEntity.ID)
					} else {
						logger.Info("generate completed for stack", "stackID", params.StackID, "time", time.Now())
						if yaml, err := yamlv2.Marshal(sp); err == nil {
							h.setRunToSuccess(newCtx, runEntity.ID, string(yaml))
						} else {
							logger.Error("Error marshalling generated spec", "error", err)
							h.setRunToFailed(newCtx, runEntity.ID)
						}
					}
				}
			}()

			// Call generate stack
			_, sp, err = h.stackManager.GenerateSpec(newCtx, params)
			if err != nil {
				logger.Error("Error generating stack", "error", err)
				return
			}
		})

		defer func() {
			if inBufferZone {
				logger.Info("The task is in the buffer zone, waiting for an available worker")
				h.setRunToQueued(ctx, runEntity.ID)
			}
		}()
		render.Render(w, r, handler.SuccessResponse(ctx, runEntity))
	}
}

// @Id				destroyStackAsync
// @Summary		Asynchronously destroy stack
// @Description	Start a run and asynchronously destroy stack resources by stack ID
// @Tags			stack
// @Produce		json
// @Param			stack_id	path		int		true	"Stack ID"
// @Param			workspace			query		string						true	"The target workspace to preview the spec in."
// @Param			force		query		bool	false	"Force the destroy even when the stack is locked. May cause concurrency issues!!!"
// @Param			dryrun		query		bool	false	"Destroy in dry-run mode"
// @Success		200			{object}	string	"Success"
// @Failure		400			{object}	error	"Bad Request"
// @Failure		401			{object}	error	"Unauthorized"
// @Failure		429			{object}	error	"Too Many Requests"
// @Failure		404			{object}	error	"Not Found"
// @Failure		500			{object}	error	"Internal Server Error"
// @Router			/api/v1/stacks/{stack_id}/destroy/async [post]
func (h *Handler) DestroyStackAsync() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Destroying stack asynchronously...", "stackID", params.StackID)

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

		runLogger := logutil.GetRunLogger(ctx)
		runLogger.Info("Starting destroying stack in StackManager ... This is a destroy run.", "runID", runEntity.ID)

		// Starts a safe goroutine using given recover handler
		inBufferZone := h.workerPool.Do(func() {
			// defer safe.HandleCrash(aciLoggingRecoverHandler(h.aciClient, &req, log))
			logger.Info("Async destroy in progress")
			newCtx, cancel := CopyToNewContextWithTimeout(ctx, constant.RunTimeOut)
			defer cancel() // make sure the context is canceled to free resources

			// update status of the run when exiting the async run
			defer func() {
				select {
				case <-newCtx.Done():
					logger.Info("destroy execution timed out", "stackID", params.StackID, "time", time.Now(), "timeout", newCtx.Err())
					h.setRunToCancelled(newCtx, runEntity.ID)
				default:
					if err != nil {
						logger.Info("destroy failed for stack", "stackID", params.StackID, "time", time.Now())
						h.setRunToFailed(newCtx, runEntity.ID)
					} else {
						logger.Info("destroy completed for stack", "stackID", params.StackID, "time", time.Now())
						h.setRunToSuccess(newCtx, runEntity.ID, "destroy completed")
					}
				}
			}()

			err = h.stackManager.DestroyStack(newCtx, params, w)
			if err != nil {
				if err == stackmanager.ErrDryrunDestroy {
					logger.Info("Dry-run mode enabled, the above resources will be destroyed if dryrun is set to false")
					return
				} else {
					logger.Error("Error destroying stack", "error", err)
					return
				}
			}
		})

		defer func() {
			if inBufferZone {
				logger.Info("The task is in the buffer zone, waiting for an available worker")
				h.setRunToQueued(ctx, runEntity.ID)
			}
		}()
		render.Render(w, r, handler.SuccessResponse(ctx, runEntity))
	}
}

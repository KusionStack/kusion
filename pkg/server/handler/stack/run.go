package stack

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/render"
	"kusionstack.io/kusion/pkg/server/handler"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

// @Id				getRun
// @Summary		Get run
// @Description	Get run information by run ID
// @Tags			run
// @Produce		json
// @Param			run	path		int			true	"Run ID"
// @Success		200	{object}	entity.Run	"Success"
// @Failure		400	{object}	error		"Bad Request"
// @Failure		401	{object}	error		"Unauthorized"
// @Failure		429	{object}	error		"Too Many Requests"
// @Failure		404	{object}	error		"Not Found"
// @Failure		500	{object}	error		"Internal Server Error"
// @Router			/api/v1/runs/{run_id} [get]
func (h *Handler) GetRun() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := runRequestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Getting run...", "runID", params.RunID)

		existingEntity, err := h.stackManager.GetRunByID(ctx, params.RunID)
		handler.HandleResult(w, r, ctx, err, existingEntity)
	}
}

// @Id				getRunResult
// @Summary		Get run result
// @Description	Get run result by run ID
// @Tags			run
// @Produce		json
// @Param			run	path		int			true	"Run ID"
// @Success		200	{object}	entity.Run	"Success"
// @Failure		400	{object}	error		"Bad Request"
// @Failure		401	{object}	error		"Unauthorized"
// @Failure		429	{object}	error		"Too Many Requests"
// @Failure		404	{object}	error		"Not Found"
// @Failure		500	{object}	error		"Internal Server Error"
// @Router			/api/v1/runs/{run_id}/result [get]
func (h *Handler) GetRunResult() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := runRequestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Getting run...", "runID", params.RunID)

		existingEntity, err := h.stackManager.GetRunByID(ctx, params.RunID)
		if err != nil {
			handler.HandleResult(w, r, ctx, err, existingEntity)
			return
		}
		var resultJSON any
		err = json.Unmarshal([]byte(existingEntity.Result), &resultJSON)
		handler.HandleResult(w, r, ctx, err, resultJSON)
	}
}

// @Id				listRun
// @Summary		List runs
// @Description	List all runs
// @Tags			stack
// @Produce		json
// @Param			projectID			query		uint			false	"ProjectID to filter runs by. Default to all"
// @Param			orgID				query		uint			false	"OrgID to filter runs by. Default to all"
// @Param			projectName			query		string			false	"ProjectName to filter runs by. Default to all"
// @Param			cloud				query		string			false	"Cloud to filter runs by. Default to all"
// @Param			env					query		string			false	"Environment to filter runs by. Default to all"
// @Success		200					{object}	[]entity.Stack	"Success"
// @Failure		400					{object}	error			"Bad Request"
// @Failure		401					{object}	error			"Unauthorized"
// @Failure		429					{object}	error			"Too Many Requests"
// @Failure		404					{object}	error			"Not Found"
// @Failure		500					{object}	error			"Internal Server Error"
// @Router			/api/v1/runs [get]
func (h *Handler) ListRuns() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Listing runs...")

		projectIDParam := r.URL.Query().Get("projectID")
		stackIDParam := r.URL.Query().Get("stackID")
		workspaceParam := r.URL.Query().Get("workspace")

		filter, err := h.stackManager.BuildRunFilter(ctx, projectIDParam, stackIDParam, workspaceParam)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		runEntities, err := h.stackManager.ListRuns(ctx, filter)
		handler.HandleResult(w, r, ctx, err, runEntities)
	}
}

// TODO: StreamRunLogs to stream logs from a run using SSE
// func StreamRunLogs(c echo.Context) error {
// 	// 设置 SSE headers
// 	c.Response().Header().Set("Content-Type", "text/event-stream")
// 	c.Response().Header().Set("Cache-Control", "no-cache")
// 	c.Response().Header().Set("Connection", "keep-alive")

// 	// id := c.Param("id")
// 	logs := []string{"log1", "log2", "log3"}

// 	for {
// 		if len(logs) > 0 {
// 			for _, logMessage := range logs {
// 				fmt.Fprintf(c.Response().Writer, "data: %s\n\n", logMessage)
// 			}
// 			logs = nil
// 			c.Response().Flush()
// 		}
// 		time.Sleep(1 * time.Second)
// 	}
// }

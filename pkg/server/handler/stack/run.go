//nolint:dupl
package stack

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/render"
	response "kusionstack.io/kusion/pkg/domain/response"
	"kusionstack.io/kusion/pkg/server/handler"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

// @Id				getRun
// @Summary		Get run
// @Description	Get run information by run ID
// @Tags			run
// @Produce		json
// @Param			runID	path		int									true	"Run ID"
// @Success		200		{object}	handler.Response{data=entity.Run}	"Success"
// @Failure		400		{object}	error								"Bad Request"
// @Failure		401		{object}	error								"Unauthorized"
// @Failure		429		{object}	error								"Too Many Requests"
// @Failure		404		{object}	error								"Not Found"
// @Failure		500		{object}	error								"Internal Server Error"
// @Router			/api/v1/runs/{runID} [get]
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
// @Param			runID	path		int							true	"Run ID"
// @Success		200		{object}	handler.Response{data=any}	"Success"
// @Failure		400		{object}	error						"Bad Request"
// @Failure		401		{object}	error						"Unauthorized"
// @Failure		429		{object}	error						"Too Many Requests"
// @Failure		404		{object}	error						"Not Found"
// @Failure		500		{object}	error						"Internal Server Error"
// @Router			/api/v1/runs/{runID}/result [get]
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
// @Param			projectID	query		uint													false	"ProjectID to filter runs by. Default to all"
// @Param			type		query		[]string												false	"RunType to filter runs by. Default to all"
// @Param			status		query		[]string												false	"RunStatus to filter runs by. Default to all"
// @Param			stackID		query		uint													false	"StackID to filter runs by. Default to all"
// @Param			workspace	query		string													false	"Workspace to filter runs by. Default to all"
// @Param			startTime	query		string													false	"StartTime to filter runs by. Default to all. Format: RFC3339"
// @Param			endTime		query		string													false	"EndTime to filter runs by. Default to all. Format: RFC3339"
// @Param			page		query		uint													false	"The current page to fetch. Default to 1"
// @Param			pageSize	query		uint													false	"The size of the page. Default to 10"
// @Success		200			{object}	handler.Response{data=response.PaginatedRunResponse}	"Success"
// @Failure		400			{object}	error													"Bad Request"
// @Failure		401			{object}	error													"Unauthorized"
// @Failure		429			{object}	error													"Too Many Requests"
// @Failure		404			{object}	error													"Not Found"
// @Failure		500			{object}	error													"Internal Server Error"
// @Router			/api/v1/runs [get]
func (h *Handler) ListRuns() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Listing runs...")

		query := r.URL.Query()
		filter, err := h.stackManager.BuildRunFilter(ctx, &query)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// List runs
		runEntities, err := h.stackManager.ListRuns(ctx, filter)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		paginatedResponse := response.PaginatedRunResponse{
			Runs:        runEntities.Runs,
			Total:       runEntities.Total,
			CurrentPage: filter.Pagination.Page,
			PageSize:    filter.Pagination.PageSize,
		}
		handler.HandleResult(w, r, ctx, err, paginatedResponse)
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

package stack

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/request"
	stackmanager "kusionstack.io/kusion/pkg/server/manager/stack"
	appmiddleware "kusionstack.io/kusion/pkg/server/middleware"

	authutil "kusionstack.io/kusion/pkg/server/util/auth"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

func (h *Handler) setRunToSuccess(ctx context.Context, runID uint, result any) {
	logger := logutil.GetLogger(ctx)
	runLogs := logutil.GetRunLoggerBuffer(ctx)
	resultBytes, err := json.Marshal(result)
	if err != nil {
		logger.Error("Error marshalling preview changes", "error", err)
		return
	}
	// Update the Run object in database to include the preview result
	updateRunResultPayload := request.UpdateRunResultRequest{
		Result: string(resultBytes),
		Status: string(constant.RunStatusSucceeded),
		Logs:   runLogs.String(),
	}
	_, err = h.stackManager.UpdateRunResultAndStatusByID(ctx, runID, updateRunResultPayload)
	if err != nil {
		logger.Error("Error updating run result after success", "error", err)
		return
	}
}

func (h *Handler) setRunToFailed(ctx context.Context, runID uint) {
	logger := logutil.GetLogger(ctx)
	runLogs := logutil.GetRunLoggerBuffer(ctx)
	logsNew := strings.ReplaceAll(runLogs.String(), "\n", "\n\n")
	updateRunResultPayload := request.UpdateRunResultRequest{
		Result: constant.RunResultFailed,
		Status: string(constant.RunStatusFailed),
		Logs:   logsNew,
	}
	_, err := h.stackManager.UpdateRunResultAndStatusByID(ctx, runID, updateRunResultPayload)
	if err != nil {
		logger.Error("Error updating run result after failure", "error", err)
	}
}

func (h *Handler) setRunToCancelled(ctx context.Context, runID uint) {
	logger := logutil.GetLogger(ctx)
	runLogs := logutil.GetRunLoggerBuffer(ctx)
	updateRunResultPayload := request.UpdateRunResultRequest{
		Result: constant.RunResultCancelled,
		Status: string(constant.RunStatusCancelled),
		Logs:   runLogs.String(),
	}
	newCtx := CopyToNewContext(ctx)
	_, err := h.stackManager.UpdateRunResultAndStatusByID(newCtx, runID, updateRunResultPayload)
	if err != nil {
		logger.Error("Error updating run result after timeout", "error", err)
	}
}

func (h *Handler) setRunToQueued(ctx context.Context, runID uint) {
	logger := logutil.GetLogger(ctx)
	runLogs := logutil.GetRunLoggerBuffer(ctx)
	updateRunResultPayload := request.UpdateRunResultRequest{
		Result: "",
		Status: string(constant.RunStatusQueued),
		Logs:   runLogs.String(),
	}
	newCtx := CopyToNewContext(ctx)
	_, err := h.stackManager.UpdateRunResultAndStatusByID(newCtx, runID, updateRunResultPayload)
	if err != nil {
		logger.Error("Error updating run result after queueing", "error", err)
	}
}

func requestHelper(r *http.Request) (context.Context, *httplog.Logger, *stackmanager.StackRequestParams, error) {
	ctx := r.Context()
	stackID := chi.URLParam(r, "stackID")
	// Get stack with repository
	id, err := strconv.Atoi(stackID)
	if err != nil {
		return ctx, nil, nil, stackmanager.ErrInvalidStackID
	}
	logger := logutil.GetLogger(ctx)
	// Get Params
	outputParam := r.URL.Query().Get("output")
	detailParam, _ := strconv.ParseBool(r.URL.Query().Get("detail"))
	dryrunParam, _ := strconv.ParseBool(r.URL.Query().Get("dryrun"))
	forceParam, _ := strconv.ParseBool(r.URL.Query().Get("force"))
	noCacheParam, _ := strconv.ParseBool(r.URL.Query().Get("noCache"))
	unlockParam, _ := strconv.ParseBool(r.URL.Query().Get("unlock"))
	watchParam, _ := strconv.ParseBool(r.URL.Query().Get("watch"))
	watchTimeoutStr := r.URL.Query().Get("watchTimeout")
	if watchTimeoutStr == "" {
		watchTimeoutStr = "120"
	}
	watchTimeoutParam, err := strconv.Atoi(watchTimeoutStr)
	if err != nil {
		return ctx, nil, nil, stackmanager.ErrInvalidWatchTimeout
	}
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
	executeParams := stackmanager.StackExecuteParams{
		Detail:              detailParam,
		Dryrun:              dryrunParam,
		Force:               forceParam,
		SpecID:              specIDParam,
		ImportResources:     importResourcesParam,
		NoCache:             noCacheParam,
		Unlock:              unlockParam,
		Watch:               watchParam,
		WatchTimeoutSeconds: watchTimeoutParam,
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

func runRequestHelper(r *http.Request) (context.Context, *httplog.Logger, *stackmanager.RunRequestParams, error) {
	ctx := r.Context()
	runID := chi.URLParam(r, "runID")
	// Get stack with repository
	id, err := strconv.Atoi(runID)
	if err != nil {
		return nil, nil, nil, stackmanager.ErrInvalidRunID
	}
	logger := logutil.GetLogger(ctx)
	params := stackmanager.RunRequestParams{
		RunID: uint(id),
	}
	return ctx, logger, &params, nil
}

func CopyToNewContext(ctx context.Context) context.Context {
	newCtx := context.Background()
	newCtx = context.WithValue(newCtx, appmiddleware.TraceIDKey, appmiddleware.GetTraceID(ctx))
	newCtx = context.WithValue(newCtx, appmiddleware.UserIDKey, appmiddleware.GetUserID(ctx))
	if logger, ok := ctx.Value(appmiddleware.APILoggerKey).(*httplog.Logger); ok {
		newCtx = context.WithValue(newCtx, appmiddleware.APILoggerKey, logger)
	}
	if runLogger, ok := ctx.Value(appmiddleware.RunLoggerKey).(*httplog.Logger); ok {
		newCtx = context.WithValue(newCtx, appmiddleware.RunLoggerKey, runLogger)
	}
	if runLoggerBuffer, ok := ctx.Value(appmiddleware.RunLoggerBufferKey).(*bytes.Buffer); ok {
		newCtx = context.WithValue(newCtx, appmiddleware.RunLoggerBufferKey, runLoggerBuffer)
	}
	return newCtx
}

func CopyToNewContextWithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	newCtx := CopyToNewContext(ctx)
	newCtxWithTimeout, cancel := context.WithTimeout(newCtx, timeout)
	return newCtxWithTimeout, cancel
}

func logStackTrace(runLogger *httplog.Logger) {
	buf := make([]byte, 1<<16) // 64KB
	stackSize := runtime.Stack(buf, true)
	runLogger.Error("Stack trace:")
	runLogger.Error(string(buf[:stackSize]))
}

type SetRunToFailedFunc func(context.Context, uint)

func handleCrash(ctx context.Context, statusHandlingFunc SetRunToFailedFunc, runID uint) {
	if r := recover(); r != nil {
		logger := logutil.GetLogger(ctx)
		runLogger := logutil.GetRunLogger(ctx)
		logger.Error("Recovered from panic during async execution:", "error", r)
		logStackTrace(logger)
		runLogger.Error("Panic recovered", "error", r)
		logStackTrace(runLogger)
		statusHandlingFunc(ctx, runID)
	}
}

func updateRunRequestPayload(requestPayload *request.CreateRunRequest, params *stackmanager.StackRequestParams, runType constant.RunType) {
	requestPayload.StackID = params.StackID
	requestPayload.Type = string(runType)
	requestPayload.Workspace = params.Workspace
}

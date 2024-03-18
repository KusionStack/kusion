package handler

import (
	"context"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	appmiddleware "kusionstack.io/kusion/pkg/server/middleware"
)

// SuccessMessage is the default success message for successful responses.
const SuccessMessage = "OK"

// Response creates a standard API response renderer.
func Response(ctx context.Context, data any, err error) render.Renderer {
	resp := &response{}

	// Set the Success and Message fields based on the error parameter.
	if err == nil {
		resp.Success = true
		resp.Message = SuccessMessage
		resp.Data = data
	} else {
		resp.Success = false
		resp.Message = err.Error()
	}

	// Include the request trace ID if available.
	if requestID := middleware.GetReqID(ctx); len(requestID) > 0 {
		resp.TraceID = requestID
	}

	// Calculate and include timing details if a start time is set.
	if startTime := appmiddleware.GetStartTime(ctx); !startTime.IsZero() {
		endTime := time.Now()
		resp.StartTime = &startTime
		resp.EndTime = &endTime
		resp.CostTime = Duration(endTime.Sub(startTime))
	}

	return resp
}

// FailureResponse creates a response renderer for a failed request.
func FailureResponse(ctx context.Context, err error) render.Renderer {
	return Response(ctx, nil, err)
}

// SuccessResponse creates a response renderer for a successful request.
func SuccessResponse(ctx context.Context, data any) render.Renderer {
	return Response(ctx, data, nil)
}

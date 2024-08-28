package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/render"
)

func HandleResult(w http.ResponseWriter, r *http.Request, ctx context.Context, err error, data any) {
	if err != nil {
		render.Render(w, r, FailureResponse(ctx, err))
		return
	}
	render.JSON(w, r, SuccessResponse(ctx, data))
}

package source

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/render"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/domain/response"
	"kusionstack.io/kusion/pkg/server/handler"
	sourcemanager "kusionstack.io/kusion/pkg/server/manager/source"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

// @Id				createSource
// @Summary		Create source
// @Description	Create a new source
// @Tags			source
// @Accept			json
// @Produce		json
// @Param			source	body		request.CreateSourceRequest				true	"Created source"
// @Success		200		{object}	handler.Response{data=entity.Source}	"Success"
// @Failure		400		{object}	error									"Bad Request"
// @Failure		401		{object}	error									"Unauthorized"
// @Failure		429		{object}	error									"Too Many Requests"
// @Failure		404		{object}	error									"Not Found"
// @Failure		500		{object}	error									"Internal Server Error"
// @Router			/api/v1/sources [post]
func (h *Handler) CreateSource() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Creating source...")

		// Decode the request body into the payload.
		var requestPayload request.CreateSourceRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Validate request payload
		if err := requestPayload.Validate(); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return created entity
		createdEntity, err := h.sourceManager.CreateSource(ctx, requestPayload)
		handler.HandleResult(w, r, ctx, err, createdEntity)
	}
}

// @Id				deleteSource
// @Summary		Delete source
// @Description	Delete specified source by ID
// @Tags			source
// @Produce		json
// @Param			sourceID	path		int								true	"Source ID"
// @Success		200			{object}	handler.Response{data=string}	"Success"
// @Failure		400			{object}	error							"Bad Request"
// @Failure		401			{object}	error							"Unauthorized"
// @Failure		429			{object}	error							"Too Many Requests"
// @Failure		404			{object}	error							"Not Found"
// @Failure		500			{object}	error							"Internal Server Error"
// @Router			/api/v1/sources/{sourceID} [delete]
func (h *Handler) DeleteSource() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Deleting source...")

		err = h.sourceManager.DeleteSourceByID(ctx, params.SourceID)
		handler.HandleResult(w, r, ctx, err, "Deletion Success")
	}
}

// @Id				updateSource
// @Summary		Update source
// @Description	Update the specified source
// @Tags			source
// @Accept			json
// @Produce		json
// @Param			sourceID	path		int										true	"Source ID"
// @Param			source		body		request.UpdateSourceRequest				true	"Updated source"
// @Success		200			{object}	handler.Response{data=entity.Source}	"Success"
// @Failure		400			{object}	error									"Bad Request"
// @Failure		401			{object}	error									"Unauthorized"
// @Failure		429			{object}	error									"Too Many Requests"
// @Failure		404			{object}	error									"Not Found"
// @Failure		500			{object}	error									"Internal Server Error"
// @Router			/api/v1/sources/{sourceID} [put]
func (h *Handler) UpdateSource() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Updating source...")

		// Decode the request body into the payload.
		var requestPayload request.UpdateSourceRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Validate request payload
		if err := requestPayload.Validate(); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return updated source
		updatedEntity, err := h.sourceManager.UpdateSourceByID(ctx, params.SourceID, requestPayload)
		handler.HandleResult(w, r, ctx, err, updatedEntity)
	}
}

// @Id				getSource
// @Summary		Get source
// @Description	Get source information by source ID
// @Tags			source
// @Produce		json
// @Param			sourceID	path		int										true	"Source ID"
// @Success		200			{object}	handler.Response{data=entity.Source}	"Success"
// @Failure		400			{object}	error									"Bad Request"
// @Failure		401			{object}	error									"Unauthorized"
// @Failure		429			{object}	error									"Too Many Requests"
// @Failure		404			{object}	error									"Not Found"
// @Failure		500			{object}	error									"Internal Server Error"
// @Router			/api/v1/sources/{sourceID} [get]
func (h *Handler) GetSource() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Getting source...")

		existingEntity, err := h.sourceManager.GetSourceByID(ctx, params.SourceID)
		handler.HandleResult(w, r, ctx, err, existingEntity)
	}
}

// @Id				listSource
// @Summary		List source
// @Description	List source information by source ID
// @Tags			source
// @Produce		json
// @Param			sourceName	query		string													false	"Source name to filter source list by. Default to all sources."
// @Param			page		query		uint													false	"The current page to fetch. Default to 1"
// @Param			pageSize	query		uint													false	"The size of the page. Default to 10"
// @Param			sortBy		query		string														false	"Which field to sort the list by. Default to id"
// @Param			ascending	query		bool														false	"Whether to sort the list in ascending order. Default to false"
// @Success		200			{object}	handler.Response{data=response.PaginatedSourceResponse}	"Success"
// @Failure		400			{object}	error													"Bad Request"
// @Failure		401			{object}	error													"Unauthorized"
// @Failure		429			{object}	error													"Too Many Requests"
// @Failure		404			{object}	error													"Not Found"
// @Failure		500			{object}	error													"Internal Server Error"
// @Router			/api/v1/sources [get]
func (h *Handler) ListSources() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Listing source...")

		// Getting source filters
		query := r.URL.Query()
		filter, sourceSortOptions, err := h.sourceManager.BuildSourceFilterAndSortOptions(ctx, &query)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// List sources with pagination.
		sourceEntities, err := h.sourceManager.ListSources(ctx, filter, sourceSortOptions)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		paginatedResponse := response.PaginatedSourceResponse{
			Sources:     sourceEntities.Sources,
			Total:       sourceEntities.Total,
			CurrentPage: filter.Pagination.Page,
			PageSize:    filter.Pagination.PageSize,
		}
		handler.HandleResult(w, r, ctx, err, paginatedResponse)
	}
}

func requestHelper(r *http.Request) (context.Context, *httplog.Logger, *SourceRequestParams, error) {
	ctx := r.Context()
	sourceID := chi.URLParam(r, "sourceID")
	// Get stack with repository
	id, err := strconv.Atoi(sourceID)
	if err != nil {
		return nil, nil, nil, sourcemanager.ErrInvalidSourceID
	}
	logger := logutil.GetLogger(ctx)
	params := SourceRequestParams{
		SourceID: uint(id),
	}
	return ctx, logger, &params, nil
}

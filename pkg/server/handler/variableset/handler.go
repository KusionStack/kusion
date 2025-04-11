package variableset

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/render"
	"k8s.io/apimachinery/pkg/labels"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/domain/response"
	"kusionstack.io/kusion/pkg/server/handler"
	"kusionstack.io/kusion/pkg/server/manager/variableset"
	"kusionstack.io/kusion/pkg/server/middleware"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

// @Id				createVariableSet
// @Summary		Create variable set
// @Description	Create a new variable set
// @Tags			variable_set
// @Accept			json
// @Produce		json
// @Param			variableSet				body		request.CreateVariableSetRequest			true	"Created variable set"
// @Success		200						{object}	handler.Response{data=entity.VariableSet}	"Success"
// @Failure		400						{object}	error										"Bad Request"
// @Failure		401						{object}	error										"Unauthorized"
// @Failure		429						{object}	error										"Too Many Requests"
// @Failure		404						{object}	error										"Not Found"
// @Failure		500						{object}	error										"Internal Server Error"
// @Router			/api/v1/variablesets 	[post]
func (h *Handler) CreateVariableSet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Creating variable set...")

		// Decode the request body into the payload.
		var requestPayload request.CreateVariableSetRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Validate request payload.
		if err := requestPayload.Validate(); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return the created entity.
		createdEntity, err := h.variableSetManager.CreateVariableSet(ctx, requestPayload)
		handler.HandleResult(w, r, ctx, err, createdEntity)
	}
}

// @Id				deleteVariableSet
// @Summary		Delete variable set
// @Description	Delete the specified variable set by name
// @Tags			variable_set
// @Produce		json
// @Param			variableSetName							path		string							true	"Variable Set Name"
// @Success		200										{object}	handler.Response{data=string}	"Success"
// @Failure		400										{object}	error							"Bad Request"
// @Failure		401										{object}	error							"Unauthorized"
// @Failure		429										{object}	error							"Too Many Requests"
// @Failure		404										{object}	error							"Not Found"
// @Failure		500										{object}	error							"Internal Server Error"
// @Router			/api/v1/variablesets/{variableSetName} 	[delete]
func (h *Handler) DeleteVariableSet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Deleting variable set...")

		err = h.variableSetManager.DeleteVariableSetByName(ctx, params.VariableSetName)
		handler.HandleResult(w, r, ctx, err, "Deletion Success")
	}
}

// @Id				updateVariableSet
// @Summary		Update variable set
// @Description	Update the specified variable set with name
// @Tags			variable_set
// @Accept			json
// @Produce		json
// @Param			variableSetName							path		string										true	"Variable Set Name"
// @Param			variableSet								body		request.UpdateVariableSetRequest			true	"Updated variable set"
// @Success		200										{object}	handler.Response{data=entity.VariableSet}	"Success"
// @Failure		400										{object}	error										"Bad Request"
// @Failure		401										{object}	error										"Unauthorized"
// @Failure		429										{object}	error										"Too Many Requests"
// @Failure		404										{object}	error										"Not Found"
// @Failure		500										{object}	error										"Internal Server Error"
// @Router			/api/v1/variablesets/{variableSetName} 																																																																																																																																																																																																																																																											[put]
func (h *Handler) UpdateVariableSet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Updating variable set...")

		// Decode the request body into the payload.
		var requestPayload request.UpdateVariableSetRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Validate request payload.
		if requestPayload.Name != "" && requestPayload.Name != params.VariableSetName {
			render.Render(w, r, handler.
				FailureResponse(ctx, errors.New("inconsistent variable set name in path and request body")))
			return
		}
		if err := requestPayload.Validate(); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return the updated variable set.
		updatedEntity, err := h.variableSetManager.UpdateVariableSetByName(ctx, params.VariableSetName, requestPayload)
		handler.HandleResult(w, r, ctx, err, updatedEntity)
	}
}

// @Id				getVariableSet
// @Summary		Get variable set
// @Description	Get variable set information by variable set name
// @Tags			variable_set
// @Produce		json
// @Param			variableSetName							path		string										true	"Variable Set Name"
// @Success		200										{object}	handler.Response{data=entity.VariableSet}	"Success"
// @Failure		400										{object}	error										"Bad Request"
// @Failure		401										{object}	error										"Unauthorized"
// @Failure		429										{object}	error										"Too Many Requests"
// @Failure		404										{object}	error										"Not Found"
// @Failure		500										{object}	error										"Internal Server Error"
// @Router			/api/v1/variablesets/{variableSetName} 																																																																																																																																																																																																																																									[get]
func (h *Handler) GetVariableSet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Getting variable set...")

		existingEntity, err := h.variableSetManager.GetVariableSetByName(ctx, params.VariableSetName)
		handler.HandleResult(w, r, ctx, err, existingEntity)
	}
}

// @Id				listVariableSets
// @Summary		List variable sets
// @Description	List variable set information
// @Tags			variable_set
// @Produce		json
// @Param			variableSetName	query		string															false	"Variable Set Name"
// @Param			page			query		uint															false	"The current page to fetch. Default to 1"
// @Param			pageSize		query		uint															false	"The size of the page. Default to 10"
// @Param			sortBy			query		string															false	"Which field to sort the list by. Default to id"
// @Param			descending		query		bool															false	"Whether to sort the list in descending order. Default to false"
// @Param			fetchAll		query		bool															false	"Whether to list all the variable sets"
// @Success		200				{object}	handler.Response{data=response.PaginatedVariableSetResponse}	"Success"
// @Failure		400				{object}	error															"Bad Request"
// @Failure		401				{object}	error															"Unauthorized"
// @Failure		429				{object}	error															"Too Many Requests"
// @Failure		404				{object}	error															"Not Found"
// @Failure		500				{object}	error															"Internal Server Error"
// @Router			/api/v1/variablesets [get]
func (h *Handler) ListVariableSets() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Listing variable sets...")

		query := r.URL.Query()

		// Get variable set filter.
		filter, variableSetSortOptions, err := h.variableSetManager.BuildVariableSetFilterAndSortOptions(ctx, &query)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// List variable sets with pagination.
		variableSetEntities, err := h.variableSetManager.ListVariableSets(ctx, filter, variableSetSortOptions)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// If the amount of variable sets exceeds the maximum result limit,
		// then indicate in the response message.
		if len(variableSetEntities.VariableSets) < variableSetEntities.Total {
			ctx = context.WithValue(ctx, middleware.ResponseMessageKey, "the result exceeds the maximum amount limit")
		}

		paginatedResponse := response.PaginatedVariableSetResponse{
			VariableSets: variableSetEntities.VariableSets,
			Total:        variableSetEntities.Total,
			CurrentPage:  filter.Pagination.Page,
			PageSize:     filter.Pagination.PageSize,
		}
		handler.HandleResult(w, r, ctx, err, paginatedResponse)
	}
}

// @Id				listVariableSetsByLabels
// @Summary		List variable sets by labels
// @Description	List variable set information by label selectors
// @Tags			variable_set
// @Produce		json
// @Param			selector	query		string															true	"Label selectors to match variable sets"
// @Success		200			{object}	handler.Response{data=response.PaginatedVariableSetResponse}	"Success"
// @Failure		400			{object}	error															"Bad Request"
// @Failure		401			{object}	error															"Unauthorized"
// @Failure		429			{object}	error															"Too Many Requests"
// @Failure		404			{object}	error															"Not Found"
// @Failure		500			{object}	error															"Internal Server Error"
// @Router			/api/v1/variablesets/matched [get]
func (h *Handler) ListVariableSetsByLabels() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Listing variable sets by labels...")

		query := r.URL.Query()

		// List variable sets by label selectors.
		selector := query.Get("selector")
		if selector == "" {
			render.Render(w, r, handler.FailureResponse(ctx, errors.New("empty label selectors")))
			return
		}

		// Fetch all the variable sets to match the label selectors.
		filter := &entity.VariableSetFilter{
			Pagination: &entity.Pagination{
				Page:     constant.CommonPageDefault,
				PageSize: constant.CommonMaxResultLimit,
			},
			FetchAll: true,
		}
		sortOptions := &entity.SortOptions{
			Field:      constant.SortByID,
			Descending: false,
		}

		variableSetEntities, err := h.variableSetManager.ListVariableSets(ctx, filter, sortOptions)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// If the amount of variable sets exceeds the maximum result limit,
		// then indicate in the response message.
		if len(variableSetEntities.VariableSets) < variableSetEntities.Total {
			ctx = context.WithValue(ctx, middleware.ResponseMessageKey, "the result exceeds the maximum amount limit")
		}

		var matchedVariableSets []*entity.VariableSet

		// Match the label selectors with Parser from `k8s.io/apimachinery/pkg/labels`.
		labelSelector, err := labels.Parse(selector)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		for _, vs := range variableSetEntities.VariableSets {
			labelSet := labels.Set(vs.Labels)
			if labelSelector.Matches(labelSet) {
				matchedVariableSets = append(matchedVariableSets, vs)
			}
		}

		selectedResponse := response.SelectedVariableSetResponse{
			VariableSets: matchedVariableSets,
			Total:        len(matchedVariableSets),
		}
		handler.HandleResult(w, r, ctx, err, selectedResponse)
	}
}

func requestHelper(r *http.Request) (context.Context, *httplog.Logger, *VariableSetRequestParams, error) {
	ctx := r.Context()
	logger := logutil.GetLogger(ctx)

	// Get URL parameters.
	variableSetName := chi.URLParam(r, "variableSetName")
	if variableSetName == "" {
		return nil, nil, nil, variableset.ErrEmptyVariableSetName
	}

	params := VariableSetRequestParams{
		VariableSetName: variableSetName,
	}

	return ctx, logger, &params, nil
}

package variable

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/render"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/domain/response"
	"kusionstack.io/kusion/pkg/server/handler"
	"kusionstack.io/kusion/pkg/server/manager/variable"
	"kusionstack.io/kusion/pkg/server/manager/variableset"
	"kusionstack.io/kusion/pkg/server/middleware"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

// @Id				createVariable
// @Summary		Create variable
// @Description	Create a new variable
// @Tags			variable
// @Accept			json
// @Produce		json
// @Param			variable			body		request.CreateVariableRequest			true	"Created variable"
// @Success		200					{object}	handler.Response{data=entity.Variable}	"Success"
// @Failure		400					{object}	error									"Bad Request"
// @Failure		401					{object}	error									"Unauthorized"
// @Failure		429					{object}	error									"Too Many Requests"
// @Failure		404					{object}	error									"Not Found"
// @Failure		500					{object}	error									"Internal Server Error"
// @Router			/api/v1/variables 	[post]
func (h *Handler) CreateVariable() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Creating variable...")

		// Decode the request body into the payload.
		var requestPayload request.CreateVariableRequest
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
		createdEntity, err := h.variableManager.CreateVariable(ctx, requestPayload)
		handler.HandleResult(w, r, ctx, err, createdEntity)
	}
}

// @Id				deleteVariable
// @Summary		Delete variable
// @Description	Delete the specified variable by name and the variable set it belongs to
// @Tags			variable
// @Produce		json
// @Param			variableSetName										path		string							true	"Variable Set Name"
// @Param			variableName										path		string							true	"Variable Name"
// @Success		200													{object}	handler.Response{data=string}	"Success"
// @Failure		400													{object}	error							"Bad Request"
// @Failure		401													{object}	error							"Unauthorized"
// @Failure		429													{object}	error							"Too Many Requests"
// @Failure		404													{object}	error							"Not Found"
// @Failure		500													{object}	error							"Internal Server Error"
// @Router			/api/v1/variables/{variableSetName}/{variableName} 	[delete]
func (h *Handler) DeleteVariable() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Deleting variable...")

		err = h.variableManager.DeleteVariableByNameAndVariableSet(ctx, params.VariableName, params.VariableSetName)
		handler.HandleResult(w, r, ctx, err, "Deletion Success")
	}
}

// @Id				updateVariable
// @Summary		Update variable
// @Description	Update the specified variable with name and the variable set it belongs to
// @Tags			variable
// @Accept			json
// @Produce		json
// @Param			variableSetName										path		string									true	"Variable Set Name"
// @Param			variableName										path		string									true	"Variable Name"
// @Param			variable											body		request.UpdateVariableRequest			true	"Updated variable"
// @Success		200													{object}	handler.Response{data=entity.Variable}	"Success"
// @Failure		400													{object}	error									"Bad Request"
// @Failure		401													{object}	error									"Unauthorized"
// @Failure		429													{object}	error									"Too Many Requests"
// @Failure		404													{object}	error									"Not Found"
// @Failure		500													{object}	error									"Internal Server Error"
// @Router			/api/v1/variables/{variableSetName}/{variableName} 																																																																																																																																																																																																																																																									[put]
func (h *Handler) UpdateVariable() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Updating variable...")

		// Decode the request body into the payload.
		var requestPayload request.UpdateVariableRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Validate request payload.
		if requestPayload.Name != "" && requestPayload.Name != params.VariableName {
			render.Render(w, r, handler.
				FailureResponse(ctx, errors.New("inconsistent variable name in path and request body")))
			return
		}
		if requestPayload.VariableSet != "" && requestPayload.VariableSet != params.VariableSetName {
			render.Render(w, r, handler.
				FailureResponse(ctx, errors.New("inconsistent variable set name in path and request body")))
			return
		}
		if err := requestPayload.Validate(); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return the updated variable.
		updatedEntity, err := h.variableManager.
			UpdateVariableByNameAndVariableSet(ctx, params.VariableName, params.VariableSetName, requestPayload)
		handler.HandleResult(w, r, ctx, err, updatedEntity)
	}
}

// @Id				getVariable
// @Summary		Get variable
// @Description	Get variable information by variable name and the variable set it belongs
// @Tags			variable
// @Produce		json
// @Param			variableSetName										path		string									true	"Variable Set Name"
// @Param			variableName										path		string									true	"Variable Name"
// @Success		200													{object}	handler.Response{data=entity.Variable}	"Success"
// @Failure		400													{object}	error									"Bad Request"
// @Failure		401													{object}	error									"Unauthorized"
// @Failure		429													{object}	error									"Too Many Requests"
// @Failure		404													{object}	error									"Not Found"
// @Failure		500													{object}	error									"Internal Server Error"
// @Router			/api/v1/variables/{variableSetName}/{variableName} 																																																																																																																																																																																																																																							[get]
func (h *Handler) GetVariable() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Getting variable...")

		existingEntity, err := h.variableManager.GetVariableByNameAndVariableSet(ctx,
			params.VariableName, params.VariableSetName)
		handler.HandleResult(w, r, ctx, err, existingEntity)
	}
}

// @Id				listVariables
// @Summary		List variables
// @Description	List variable information
// @Tags			variable
// @Produce		json
// @Param			variableName	query		string														false	"Variable Name"
// @Param			variableSetName	query		string														false	"Variable Set Name"
// @Param			page			query		uint														false	"The current page to fetch. Default to 1"
// @Param			pageSize		query		uint														false	"The size of the page. Default to 10"
// @Param			sortBy			query		string														false	"Which field to sort the list by. Default to id"
// @Param			descending		query		bool														false	"Whether to sort the list in descending order. Default to false"
// @Param			fetchAll		query		bool														false	"Whether to list all the variables"
// @Success		200				{object}	handler.Response{data=response.PaginatedVariableResponse}	"Success"
// @Failure		400				{object}	error														"Bad Request"
// @Failure		401				{object}	error														"Unauthorized"
// @Failure		429				{object}	error														"Too Many Requests"
// @Failure		404				{object}	error														"Not Found"
// @Failure		500				{object}	error														"Internal Server Error"
// @Router			/api/v1/variables [get]
func (h *Handler) ListVariables() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Listing variables...")

		query := r.URL.Query()

		// Get variable filter.
		filter, variableSortOptions, err := h.variableManager.BuildVariableFilterAndSortOptions(ctx, &query)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// List variables with pagination.
		variableEntities, err := h.variableManager.ListVariables(ctx, filter, variableSortOptions)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// If the amount of variable sets exceeds the maximum result limit,
		// then indicate in the response message.
		if len(variableEntities.Variables) < variableEntities.Total {
			ctx = context.WithValue(ctx, middleware.ResponseMessageKey, "the result exceeds the maximum amount limit")
		}

		paginatedResponse := response.PaginatedVariableResponse{
			Variables:   variableEntities.Variables,
			Total:       variableEntities.Total,
			CurrentPage: filter.Pagination.Page,
			PageSize:    filter.Pagination.PageSize,
		}
		handler.HandleResult(w, r, ctx, err, paginatedResponse)
	}
}

func requestHelper(r *http.Request) (context.Context, *httplog.Logger, *VariableRequestParams, error) {
	ctx := r.Context()
	logger := logutil.GetLogger(ctx)

	// Get URL parameters.
	variableSetName := chi.URLParam(r, "variableSetName")
	if variableSetName == "" {
		return nil, nil, nil, variableset.ErrEmptyVariableSetName
	}

	variableName := chi.URLParam(r, "variableName")
	if variableName == "" {
		return nil, nil, nil, variable.ErrEmptyVariableName
	}

	params := VariableRequestParams{
		VariableSetName: variableSetName,
		VariableName:    variableName,
	}

	return ctx, logger, &params, nil
}

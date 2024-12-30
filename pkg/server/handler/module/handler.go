package module

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
	modulemanager "kusionstack.io/kusion/pkg/server/manager/module"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

// @Id				createModule
// @Summary		Create module
// @Description	Create a new Kusion module
// @Tags			module
// @Accept			json
// @Produce		json
// @Param			module				body		request.CreateModuleRequest				true	"Created module"
// @Success		200					{object}	handler.Response{data=entity.Module}	"Success"
// @Failure		400					{object}	error									"Bad Request"
// @Failure		401					{object}	error									"Unauthorized"
// @Failure		429					{object}	error									"Too Many Requests"
// @Failure		404					{object}	error									"Not Found"
// @Failure		500					{object}	error									"Internal Server Error"
// @Router			/api/v1/modules 	[post]
func (h *Handler) CreateModule() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Creating module...")

		// Decode the request body into the payload.
		var requestPayload request.CreateModuleRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return the created entity.
		createdEntity, err := h.moduleManager.CreateModule(ctx, requestPayload)
		handler.HandleResult(w, r, ctx, err, createdEntity)
	}
}

// @Id				deleteModule
// @Summary		Delete module
// @Description	Delete the specified module by name
// @Tags			module
// @Produce		json
// @Param			name					path		string							true	"Module Name"
// @Success		200						{object}	handler.Response{data=string}	"Success"
// @Failure		400						{object}	error							"Bad Request"
// @Failure		401						{object}	error							"Unauthorized"
// @Failure		429						{object}	error							"Too Many Requests"
// @Failure		404						{object}	error							"Not Found"
// @Failure		500						{object}	error							"Internal Server Error"
// @Router			/api/v1/modules/{name} 	[delete]
func (h *Handler) DeleteModule() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Deleting module...")

		err = h.moduleManager.DeleteModuleByName(ctx, params.ModuleName)
		handler.HandleResult(w, r, ctx, err, "Deletion Success")
	}
}

// @Id				updateModule
// @Summary		Update module
// @Description	Update the specified module
// @Tags			module
// @Accept			json
// @Produce		json
// @Param			name					path		string									true	"Module Name"
// @Param			module					body		request.UpdateModuleRequest				true	"Updated module"
// @Success		200						{object}	handler.Response{data=entity.Module}	"Success"
// @Failure		400						{object}	error									"Bad Request"
// @Failure		401						{object}	error									"Unauthorized"
// @Failure		429						{object}	error									"Too Many Requests"
// @Failure		404						{object}	error									"Not Found"
// @Failure		500						{object}	error									"Internal Server Error"
// @Router			/api/v1/modules/{moduleName} 																																																																																																																																																																	[put]
func (h *Handler) UpdateModule() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Updating module...")

		// Decode the request body into the payload.
		var requestPayload request.UpdateModuleRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return the updated module.
		updatedEntity, err := h.moduleManager.UpdateModuleByName(ctx, params.ModuleName, requestPayload)
		handler.HandleResult(w, r, ctx, err, updatedEntity)
	}
}

// @Id				getModule
// @Summary		Get module
// @Description	Get module information by module name
// @Tags			module
// @Produce		json
// @Param			name					path		string									true	"Module Name"
// @Success		200						{object}	handler.Response{data=entity.Module}	"Success"
// @Failure		400						{object}	error									"Bad Request"
// @Failure		401						{object}	error									"Unauthorized"
// @Failure		429						{object}	error									"Too Many Requests"
// @Failure		404						{object}	error									"Not Found"
// @Failure		500						{object}	error									"Internal Server Error"
// @Router			/api/v1/modules/{name} 																																																																																																																																															[get]
func (h *Handler) GetModule() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Getting module...")

		existingEntity, err := h.moduleManager.GetModuleByName(ctx, params.ModuleName)
		handler.HandleResult(w, r, ctx, err, existingEntity)
	}
}

// @Id				listModule
// @Summary		List module
// @Description	List module information
// @Tags			module
// @Produce		json
// @Param			workspaceID	query		uint													false	"Workspace ID to filter module list by. Default to all workspaces."
// @Param			moduleName	query		string													false	"Module name to filter module list by. Default to all modules."
// @Param			page		query		uint													false	"The current page to fetch. Default to 1"
// @Param			pageSize	query		uint													false	"The size of the page. Default to 10"
// @Success		200			{object}	handler.Response{data=response.PaginatedModuleResponse}	"Success"
// @Failure		400			{object}	error													"Bad Request"
// @Failure		401			{object}	error													"Unauthorized"
// @Failure		429			{object}	error													"Too Many Requests"
// @Failure		404			{object}	error													"Not Found"
// @Failure		500			{object}	error													"Internal Server Error"
// @Router			/api/v1/modules [get]
func (h *Handler) ListModules() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Listing module...")

		query := r.URL.Query()

		// Get module filter.
		filter, err := h.moduleManager.BuildModuleFilter(ctx, &query)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// List modules with pagination.
		wsIDParam := query.Get("workspaceID")
		if wsIDParam != "" {
			wsID, err := strconv.Atoi(wsIDParam)
			if err != nil {
				render.Render(w, r, handler.FailureResponse(ctx, modulemanager.ErrInvalidWorkspaceID))
				return
			}

			// List modules in the specified workspace with pagination.
			moduleEntities, err := h.moduleManager.ListModulesByWorkspaceID(ctx, uint(wsID), filter)
			if err != nil {
				render.Render(w, r, handler.FailureResponse(ctx, err))
				return
			}

			paginatedResponse := response.PaginatedModuleResponse{
				ModulesWithVersion: moduleEntities.ModulesWithVersion,
				Total:              moduleEntities.Total,
				CurrentPage:        filter.Pagination.Page,
				PageSize:           filter.Pagination.PageSize,
			}
			handler.HandleResult(w, r, ctx, err, paginatedResponse)
			return
		}

		moduleEntities, err := h.moduleManager.ListModules(ctx, filter)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		paginatedResponse := response.PaginatedModuleResponse{
			Modules:     moduleEntities.Modules,
			Total:       moduleEntities.Total,
			CurrentPage: filter.Pagination.Page,
			PageSize:    filter.Pagination.PageSize,
		}
		handler.HandleResult(w, r, ctx, err, paginatedResponse)
	}
}

func requestHelper(r *http.Request) (context.Context, *httplog.Logger, *ModuleRequestParams, error) {
	ctx := r.Context()
	logger := logutil.GetLogger(ctx)

	moduleName := chi.URLParam(r, "moduleName")
	if moduleName == "" {
		return nil, nil, nil, modulemanager.ErrEmptyModuleName
	}

	params := ModuleRequestParams{
		ModuleName: moduleName,
	}

	return ctx, logger, &params, nil
}

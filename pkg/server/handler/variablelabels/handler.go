package variablelabels

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/render"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/server/handler"
	variablelabelsmanager "kusionstack.io/kusion/pkg/server/manager/variablelabels"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

// @Id				createVariableLabels
// @Summary		Create variable labels
// @Description	Create a new set of variable labels
// @Tags			variable_labels
// @Accept			json
// @Produce		json
// @Param			variable_labels				body		request.CreateVariableLabelsRequest	true	"Created variable labels"
// @Success		200							{object}	entity.VariableLabels				"Success"
// @Failure		400							{object}	error								"Bad Request"
// @Failure		401							{object}	error								"Unauthorized"
// @Failure		429							{object}	error								"Too Many Requests"
// @Failure		404							{object}	error								"Not Found"
// @Failure		500							{object}	error								"Internal Server Error"
// @Router			/api/v1/variable-labels 	[post]
func (h *Handler) CreateVariableLabels() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Creating variable with labels...")

		// Decode the request body into the payload.
		var requestPayload request.CreateVariableLabelsRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return the created entity.
		createdEntity, err := h.variableLabelsManager.CreateVariableLabels(ctx, requestPayload)
		handler.HandleResult(w, r, ctx, err, createdEntity)
	}
}

// @Id				deleteVariableLabels
// @Summary		Delete variable labels
// @Description	Delete a specified variable with labels
// @Tags			variable_labels
// @Produce		json
// @Param			key								path		string	true	"Variable Key"
// @Success		200								{object}	string	"Success"
// @Failure		400								{object}	error	"Bad Request"
// @Failure		401								{object}	error	"Unauthorized"
// @Failure		429								{object}	error	"Too Many Requests"
// @Failure		404								{object}	error	"Not Found"
// @Failure		500								{object}	error	"Internal Server Error"
// @Router			/api/v1/variable-labels/{key} 	[delete]
func (h *Handler) DeleteVariableLabels() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Deleting variable labels...")

		err = h.variableLabelsManager.DeleteVariableLabelsByKey(ctx, params.Key)
		handler.HandleResult(w, r, ctx, err, "Deletion Success")
	}
}

// @Id				updateVariableLabels
// @Summary		Update variable labels
// @Description	Update the specified variable labels
// @Tags			variable_labels
// @Accept			json
// @Produce		json
// @Param			key								path		string								true	"Variable Key"
// @Param			variable_labels					body		request.UpdateVariableLabelsRequest	true	"Updated Variable Labels"
// @Success		200								{object}	entity.VariableLabels				"Success"
// @Failure		400								{object}	error								"Bad Request"
// @Failure		401								{object}	error								"Unauthorized"
// @Failure		429								{object}	error								"Too Many Requests"
// @Failure		404								{object}	error								"Not Found"
// @Failure		500								{object}	error								"Internal Server Error"
// @Router			/api/v1/variable-labels/{key} 																																																																																											[put]
func (h *Handler) UpdateVariableLabels() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Updating variable labels...")

		// Decode the request body into the payload.
		var requestPayload request.UpdateVariableLabelsRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Return the updated variable labels.
		updatedEntity, err := h.variableLabelsManager.UpdateVariableLabels(ctx, params.Key, requestPayload)
		handler.HandleResult(w, r, ctx, err, updatedEntity)
	}
}

// @Id				getVariableLabels
// @Summary		Get variable labels
// @Description	Get variable labels by variable key
// @Tags			variable_labels
// @Produce		json
// @Param			key								path		string					true	"Variable Key"
// @Success		200								{object}	entity.VariableLabels	"Success"
// @Failure		400								{object}	error					"Bad Request"
// @Failure		401								{object}	error					"Unauthorized"
// @Failure		429								{object}	error					"Too Many Requests"
// @Failure		404								{object}	error					"Not Found"
// @Failure		500								{object}	error					"Internal Server Error"
// @Router			/api/v1/variable-labels/{key} 																																																																[get]
func (h *Handler) GetVariableLabels() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Getting variable labels...")

		existingEntity, err := h.variableLabelsManager.GetVariableLabelsByKey(ctx, params.Key)
		handler.HandleResult(w, r, ctx, err, existingEntity)
	}
}

// @Id				listVariableLabels
// @Summary		List variable labels
// @Description	List variable labels
// @Tags			variable_labels
// @Produce		json
// @Param			labels		query		[]string							false	"Variable labels to filter variables by. Default to all(empty) labels."
// @Param			page		query		string								false	"Page number of the paginated list results. Default to 1."
// @Param			pageSize	query		string								false	"Page size of the paginated list results. If not set, the result won't be paginated."
// @Success		200			{object}	[]entity.VariableLabelsListResult	"Success"
// @Failure		400			{object}	error								"Bad Request"
// @Failure		401			{object}	error								"Unauthorized"
// @Failure		429			{object}	error								"Too Many Requests"
// @Failure		404			{object}	error								"Not Found"
// @Failure		500			{object}	error								"Internal Server Error"
// @Router			/api/v1/variable-labels [get]
func (h *Handler) ListVariableLabels() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Listing variable labels...")

		query := r.URL.Query()
		labels := strings.Split(query.Get("labels"), ",")
		page, _ := strconv.Atoi(query.Get("page"))
		if page <= 0 {
			page = constant.RunPageDefault
		}
		pageSize, _ := strconv.Atoi(query.Get("pageSize"))
		if pageSize <= 0 {
			// Set `pageSize` to 0 will get an un-paginated list result here.
			pageSize = 0
		}

		filter := &entity.VariableLabelsFilter{
			Labels: labels,
			Pagination: &entity.Pagination{
				Page:     page,
				PageSize: pageSize,
			},
		}

		// List the variable labels.
		variableLabelsListResult, err := h.variableLabelsManager.ListVariableLabels(ctx, filter)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		handler.HandleResult(w, r, ctx, err, variableLabelsListResult)
	}
}

func requestHelper(r *http.Request) (context.Context, *httplog.Logger, *VariableLabelsParams, error) {
	ctx := r.Context()
	logger := logutil.GetLogger(ctx)

	variableKey := chi.URLParam(r, "variableKey")
	if variableKey == "" {
		return nil, nil, nil, variablelabelsmanager.ErrEmptyVariableKey
	}

	params := VariableLabelsParams{
		Key: variableKey,
	}

	return ctx, logger, &params, nil
}

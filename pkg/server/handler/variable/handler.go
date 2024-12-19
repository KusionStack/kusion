package variable

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/render"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/server/handler"
	"kusionstack.io/kusion/pkg/server/manager/variable"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

// @Id				createVariable
// @Summary		Create variable
// @Description	Create a new variable
// @Tags			variable
// @Accept			json
// @Produce		json
// @Param			variable			body		request.CreateVariableSetRequest	true	"Created variable"
// @Success		200					{object}	entity.Variable						"Success"
// @Failure		400					{object}	error								"Bad Request"
// @Failure		401					{object}	error								"Unauthorized"
// @Failure		429					{object}	error								"Too Many Requests"
// @Failure		404					{object}	error								"Not Found"
// @Failure		500					{object}	error								"Internal Server Error"
// @Router			/api/v1/variables 	[post]
func (h *Handler) CreateVariable() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Creating variable...")

		// Decode the request body into the payload.
		var requestPayload request.CreateVariableSetRequest
		if err := requestPayload.Decode(r); err != nil {
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
// @Description	Delete a specified variable with fqn
// @Tags			variable
// @Produce		json
// @Param			fqn							path		string	true	"Variable Fqn"
// @Success		200							{object}	string	"Success"
// @Failure		400							{object}	error	"Bad Request"
// @Failure		401							{object}	error	"Unauthorized"
// @Failure		429							{object}	error	"Too Many Requests"
// @Failure		404							{object}	error	"Not Found"
// @Failure		500							{object}	error	"Internal Server Error"
// @Router			/api/v1/variables/{fqn} 	[delete]
func (h *Handler) DeleteVariable() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Deleting variable...")

		err = h.variableManager.DeleteVariableByFqn(ctx, params.Fqn)
		handler.HandleResult(w, r, ctx, err, "Deletion Success")
	}
}

// @Id				updateVariable
// @Summary		Update variable
// @Description	Update the specified variable
// @Tags			variable
// @Accept			json
// @Produce		json
// @Param			fqn							path		string							true	"Variable Fqn"
// @Param			variable					body		request.UpdateVariableRequest	true	"Updated Variable"
// @Success		200							{object}	entity.Variable					"Success"
// @Failure		400							{object}	error							"Bad Request"
// @Failure		401							{object}	error							"Unauthorized"
// @Failure		429							{object}	error							"Too Many Requests"
// @Failure		404							{object}	error							"Not Found"
// @Failure		500							{object}	error							"Internal Server Error"
// @Router			/api/v1/variables/{fqn} 																																																																																			[put]
func (h *Handler) UpdateVariable() http.HandlerFunc {
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

		// Return the updated variable labels.
		updatedEntity, err := h.variableManager.UpdateVariable(ctx, params.Fqn, requestPayload)
		handler.HandleResult(w, r, ctx, err, updatedEntity)
	}
}

// @Id				getVariable
// @Summary		Get variable
// @Description	Get variable by variable fqn
// @Tags			variable
// @Produce		json
// @Param			fqn							path		string			true	"Variable Fqn"
// @Success		200							{object}	entity.Variable	"Success"
// @Failure		400							{object}	error			"Bad Request"
// @Failure		401							{object}	error			"Unauthorized"
// @Failure		429							{object}	error			"Too Many Requests"
// @Failure		404							{object}	error			"Not Found"
// @Failure		500							{object}	error			"Internal Server Error"
// @Router			/api/v1/variables/{fqn} 																																																[get]
func (h *Handler) GetVariable() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx, logger, params, err := requestHelper(r)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		logger.Info("Getting variable set...")

		existingEntity, err := h.variableManager.GetVariableByFqn(ctx, params.Fqn)
		handler.HandleResult(w, r, ctx, err, existingEntity)
	}
}

// @Id				listVariables
// @Summary		List variables
// @Description	List variables
// @Tags			variable
// @Produce		json
// @Param			variable	body		request.ListVariableSetRequest	true	"Variable labels to filter variables by."
// @Success		200			{object}	entity.VariableLabelsListResult	"Success"
// @Failure		400			{object}	error							"Bad Request"
// @Failure		401			{object}	error							"Unauthorized"
// @Failure		429			{object}	error							"Too Many Requests"
// @Failure		404			{object}	error							"Not Found"
// @Failure		500			{object}	error							"Internal Server Error"
// @Router			/api/v1/variables [get]
func (h *Handler) ListVariables() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context.
		ctx := r.Context()
		logger := logutil.GetLogger(ctx)
		logger.Info("Listing variables with specified labels...")

		// Decode the request body into the payload.
		var requestPayload request.ListVariableSetRequest
		if err := requestPayload.Decode(r); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// First, get the variables in the scope of the specified labels.
		var labelScope []string
		for labelKey := range requestPayload.Labels {
			labelScope = append(labelScope, labelKey)
		}

		variableLabelsListResult, err := h.variableLabelsManager.ListVariableLabels(ctx, &entity.VariableLabelsFilter{
			Labels: labelScope,
			Pagination: &entity.Pagination{
				// Here we don't need to get the paginated results.
				PageSize: 0,
			},
		})
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Second, get the corresponding variable records, and calculate the label score
		// of the variable.
		var variableRes []*entity.Variable
		for _, variableLabel := range variableLabelsListResult.VariableLabels {
			key := variableLabel.VariableKey
			variableListResult, err := h.variableManager.ListVariable(ctx, &entity.VariableFilter{
				Key: key,
				Pagination: &entity.Pagination{
					// Here we don't need to get the paginated results.
					PageSize: 0,
				},
			})
			if err != nil {
				render.Render(w, r, handler.FailureResponse(ctx, err))
				return
			}

			score := -1
			var tmpVariable *entity.Variable
			for _, variable := range variableListResult.Variables {
				currentScore := CalculateLabelMatchScore(variable, variableLabel, requestPayload.Labels)
				if score < currentScore {
					score = currentScore
					tmpVariable = variable
				}
			}

			if score > 0 {
				variableRes = append(variableRes, tmpVariable)
			}
		}

		// Finally, return the variables with the highest matching score.
		variableListResult := &entity.VariableListResult{
			Variables: variableRes,
			Total:     len(variableRes),
		}
		handler.HandleResult(w, r, ctx, err, variableListResult)
	}
}

func requestHelper(r *http.Request) (context.Context, *httplog.Logger, *VariableParams, error) {
	ctx := r.Context()
	logger := logutil.GetLogger(ctx)

	variableFqn := chi.URLParam(r, "variableFqn")
	if variableFqn == "" {
		return nil, nil, nil, variable.ErrEmptyVariableFqn
	}

	params := VariableParams{
		Fqn: variableFqn,
	}

	return ctx, logger, &params, nil
}

package stack

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/backend"
	engineapi "kusionstack.io/kusion/pkg/engine/api"
	buildersapi "kusionstack.io/kusion/pkg/engine/api/builders"
	"kusionstack.io/kusion/pkg/server/handler"
	"kusionstack.io/kusion/pkg/server/util"
)

func ExecutePreview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the context and logger from the request.
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Executing Preview...")
		projectParam := chi.URLParam(r, "projectName")
		stackParam := chi.URLParam(r, "stackName")
		workspaceParam := chi.URLParam(r, "workspaceName")
		formatParam := r.URL.Query().Get("output")
		// TODO: This is temporary. The path should be looked up based on project and stack eventually
		pathParam := r.URL.Query().Get("path")
		// TODO: Define default behaviors
		detailParam, _ := strconv.ParseBool(r.URL.Query().Get("detail"))
		isKCLPackageParam, _ := strconv.ParseBool(r.URL.Query().Get("isKCLPackage"))

		// Build API inputs
		intentOptions, previewOptions, project, stack := buildOptions(projectParam, stackParam, pathParam, isKCLPackageParam)

		// Generate spec
		sp, err := engineapi.Intent(intentOptions, project, stack)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// return immediately if no resource found in stack
		// todo: if there is no resource, should still do diff job; for now, if output is json format, there is no hint
		if sp == nil || len(sp.Resources) == 0 {
			if formatParam != engineapi.JSONOutput {
				logger.Info("No resource change found in this stack...")
				render.Render(w, r, handler.SuccessResponse(ctx, "No resource change found in this stack."))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Get state storage from cli backend options, environment variables, workspace backend configs
		// TODO: Backend should be looked up based on project and stack
		backendInstance, err := backend.NewBackend("")
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		stateStorage := backendInstance.StateStorage(projectParam, stackParam, pathParam)

		// Compute changes for preview
		changes, err := engineapi.Preview(previewOptions, stateStorage, sp, project, stack, workspaceParam)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// If output format is json, return details without any summary or formatting
		if formatParam == engineapi.JSONOutput {
			var previewChanges []byte
			previewChanges, err = json.Marshal(changes)
			if err != nil {
				render.Render(w, r, handler.FailureResponse(ctx, err))
				return
			}
			logger.Info(string(previewChanges))
			render.Render(w, r, handler.SuccessResponse(ctx, string(previewChanges)))
			return
		}

		if changes.AllUnChange() {
			logger.Info("All resources are reconciled. No diff found")
			render.Render(w, r, handler.SuccessResponse(ctx, "All resources are reconciled. No diff found"))
			return
		}

		// Summary preview table
		changes.Summary(w, true)

		// Detail detection
		if detailParam {
			render.Render(w, r, handler.SuccessResponse(ctx, changes.Diffs(true)))
		}

	}
}

func buildOptions(projectParam, stackParam, pathParam string, isKCLPackageParam bool) (*buildersapi.Options, *engineapi.APIOptions, *v1.Project, *v1.Stack) {
	// Construct intent options
	intentOptions := &buildersapi.Options{
		IsKclPkg:  isKCLPackageParam,
		WorkDir:   pathParam + "/dev",
		Arguments: map[string]string{},
		NoStyle:   true,
	}
	project := &v1.Project{
		Name: projectParam,
		Path: pathParam,
	}
	stack := &v1.Stack{
		Name: stackParam,
		Path: pathParam + "/" + stackParam,
	}
	// Construct preview api option
	// TODO: Complete preview options
	// TODO: Operator should be derived from auth info
	// TODO: Cluster should be derived from workspace config
	previewOptions := &engineapi.APIOptions{
		// Operator:     "operator",
		// Cluster:      "cluster",
		// IgnoreFields: []string{},
	}
	return intentOptions, previewOptions, project, stack
}

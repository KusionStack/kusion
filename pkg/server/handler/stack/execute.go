package stack

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	yamlv2 "gopkg.in/yaml.v2"
	"gorm.io/gorm"
	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/domain/constant"
	engineapi "kusionstack.io/kusion/pkg/engine/api"
	sourceapi "kusionstack.io/kusion/pkg/engine/api/source"
	"kusionstack.io/kusion/pkg/server/handler"
	"kusionstack.io/kusion/pkg/server/util"
)

// @Summary      Preview stack
// @Description  Preview stack information by stack ID
// @Produce      json
// @Param        id   path      int                 true  "Stack ID"
// @Success      200  {object}  entity.Stack       "Success"
// @Failure      400  {object}  errors.DetailError  "Bad Request"
// @Failure      401  {object}  errors.DetailError  "Unauthorized"
// @Failure      429  {object}  errors.DetailError  "Too Many Requests"
// @Failure      404  {object}  errors.DetailError  "Not Found"
// @Failure      500  {object}  errors.DetailError  "Internal Server Error"
// @Router       /api/v1/stack/{stackID}/preview [post]
func (h *Handler) PreviewStack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Previewing stack...")
		// Get params from URL parameter
		stackID := chi.URLParam(r, "stackID")

		// Get params from query parameter
		formatParam := r.URL.Query().Get("output")
		// TODO: Define default behaviors
		detailParam, _ := strconv.ParseBool(r.URL.Query().Get("detail"))
		// kpmParam, _ := strconv.ParseBool(r.URL.Query().Get("kpm"))
		// TODO: Should match automatically eventually
		workspaceParam := r.URL.Query().Get("workspace")

		// Get stack with repository
		id, err := strconv.Atoi(stackID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidStacktID))
			return
		}
		stackEntity, err := h.stackRepo.Get(ctx, uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrGettingNonExistingStack))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Get project by id
		project, err := stackEntity.Project.ConvertToCore()
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Get stack by id
		stack, err := stackEntity.ConvertToCore()
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// get workspace configurations
		bk, err := backend.NewBackend("")
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		wsStorage, err := bk.WorkspaceStorage()
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		ws, err := wsStorage.Get(workspaceParam)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Build API inputs
		// get project to get source and workdir
		projectEntity, err := handler.GetProjectByID(ctx, h.projectRepo, stackEntity.Project.ID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		directory, workDir, err := getWorkDirFromSource(ctx, stackEntity, projectEntity)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		previewOptions := buildOptions(false)
		stack.Path = workDir

		// Cleanup
		defer sourceapi.Cleanup(ctx, directory)

		// Generate spec
		sp, err := engineapi.GenerateSpecWithSpinner(project, stack, ws, true)
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

		// Compute state storage
		// TODO: this local storage is temporary, will support remote later
		stateStorage := bk.StateStorage(project.Name, stack.Name, ws.Name)
		logger.Info("Local state storage found", "Path", stateStorage)

		// Compute changes for preview
		changes, err := engineapi.Preview(previewOptions, stateStorage, sp, project, stack)
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

// @Summary      Build stack
// @Description  Build stack information by stack ID
// @Produce      json
// @Param        id   path      int                 true  "Stack ID"
// @Success      200  {object}  entity.Stack       "Success"
// @Failure      400  {object}  errors.DetailError  "Bad Request"
// @Failure      401  {object}  errors.DetailError  "Unauthorized"
// @Failure      429  {object}  errors.DetailError  "Too Many Requests"
// @Failure      404  {object}  errors.DetailError  "Not Found"
// @Failure      500  {object}  errors.DetailError  "Internal Server Error"
// @Router       /api/v1/stack/{stackID}/build [post]
func (h *Handler) BuildStack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Building stack...")
		// Get params from URL parameter
		stackID := chi.URLParam(r, "stackID")
		// TODO: Define default behaviors
		// kpmParam, _ := strconv.ParseBool(r.URL.Query().Get("kpm"))
		// TODO: Should match automatically eventually
		workspaceParam := r.URL.Query().Get("workspace")

		// Get stack with repository
		id, err := strconv.Atoi(stackID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidStacktID))
			return
		}
		stackEntity, err := h.stackRepo.Get(ctx, uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrGettingNonExistingStack))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Get project by id
		project, err := stackEntity.Project.ConvertToCore()
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Get stack by id
		stack, err := stackEntity.ConvertToCore()
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// get workspace configurations
		bk, err := backend.NewBackend("")
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		wsStorage, err := bk.WorkspaceStorage()
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		ws, err := wsStorage.Get(workspaceParam)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Build API inputs
		// get project to get source and workdir
		projectEntity, err := handler.GetProjectByID(ctx, h.projectRepo, stackEntity.Project.ID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		directory, workDir, err := getWorkDirFromSource(ctx, stackEntity, projectEntity)
		logger.Info("workDir derived", "workDir", workDir)
		logger.Info("directory derived", "directory", directory)

		stack.Path = workDir
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		// intentOptions, _ := buildOptions(workDir, kpmParam, false)
		// Cleanup
		defer sourceapi.Cleanup(ctx, directory)

		// Generate spec
		sp, err := engineapi.GenerateSpecWithSpinner(project, stack, ws, true)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		yaml, err := yamlv2.Marshal(sp)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		handler.HandleResult(w, r, ctx, err, string(yaml))
	}
}

// @Summary      Apply stack
// @Description  Apply stack information by stack ID
// @Produce      json
// @Param        id   path      int                 true  "Stack ID"
// @Success      200  {object}  entity.Stack       "Success"
// @Failure      400  {object}  errors.DetailError  "Bad Request"
// @Failure      401  {object}  errors.DetailError  "Unauthorized"
// @Failure      429  {object}  errors.DetailError  "Too Many Requests"
// @Failure      404  {object}  errors.DetailError  "Not Found"
// @Failure      500  {object}  errors.DetailError  "Internal Server Error"
// @Router       /api/v1/stack/{stackID}/apply [post]
func (h *Handler) ApplyStack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Applying stack...")
		// Get params from URL parameter
		stackID := chi.URLParam(r, "stackID")

		// Get params from query parameter
		formatParam := r.URL.Query().Get("output")
		dryRunParam, _ := strconv.ParseBool(r.URL.Query().Get("dryrun"))
		// TODO: Define default behaviors
		detailParam, _ := strconv.ParseBool(r.URL.Query().Get("detail"))
		// kpmParam, _ := strconv.ParseBool(r.URL.Query().Get("kpm"))
		// TODO: Should match automatically eventually
		workspaceParam := r.URL.Query().Get("workspace")

		// Get stack with repository
		id, err := strconv.Atoi(stackID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidStacktID))
			return
		}
		stackEntity, err := h.stackRepo.Get(ctx, uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrGettingNonExistingStack))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Get project by id
		project, err := stackEntity.Project.ConvertToCore()
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Get stack by id
		stack, err := stackEntity.ConvertToCore()
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// get workspace configurations
		localBackend, err := backend.NewBackend("")
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		// wsStorage, err := bk.WorkspaceStorage()
		// if err != nil {
		// 	render.Render(w, r, handler.FailureResponse(ctx, err))
		// 	return
		// }
		// ws, err := wsStorage.Get(workspaceParam)
		// if err != nil {
		// 	render.Render(w, r, handler.FailureResponse(ctx, err))
		// 	return
		// }

		// // Get backend by id
		// workspaceEntity, err := h.workspaceRepo.GetByName(ctx, workspaceParam)
		// if err != nil && err == gorm.ErrRecordNotFound {
		// 	render.Render(w, r, handler.FailureResponse(ctx, ErrWorkspaceNotFound))
		// 	return
		// } else if err != nil {
		// 	render.Render(w, r, handler.FailureResponse(ctx, err))
		// 	return
		// }
		// // Generate backend from entity
		// remoteBackend, err := NewBackendFromEntity(*workspaceEntity.Backend)
		// if err != nil {
		// 	render.Render(w, r, handler.FailureResponse(ctx, err))
		// 	return
		// }

		remoteBackend, err := h.GetBackendFromWorkspaceName(ctx, workspaceParam)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Get workspace configurations from backend
		// TODO: temporarily local for now, should be replaced by variable sets
		wsStorage, err := localBackend.WorkspaceStorage()
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		ws, err := wsStorage.Get(workspaceParam)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Build API inputs
		// get project to get source and workdir
		projectEntity, err := handler.GetProjectByID(ctx, h.projectRepo, stackEntity.Project.ID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		directory, workDir, err := getWorkDirFromSource(ctx, stackEntity, projectEntity)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		// Cleanup
		defer sourceapi.Cleanup(ctx, directory)

		executeOptions := buildOptions(dryRunParam)
		stack.Path = workDir

		// Generate spec
		sp, err := engineapi.GenerateSpecWithSpinner(project, stack, ws, true)
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

		// Compute state storage
		// TODO: this local storage is temporary, will support remote later
		stateStorage := remoteBackend.StateStorage(project.Name, stack.Name, workspaceParam)
		logger.Info("Remote state storage found", "Remote", stateStorage)
		// logger.Info("Local state storage found", "Path", stateStorage)

		// Compute changes for preview
		changes, err := engineapi.Preview(executeOptions, stateStorage, sp, project, stack)
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
		// detail detection
		if detailParam {
			changes.OutputDiff("all")
		}

		logger.Info("Start applying diffs ...")
		if err = engineapi.Apply(executeOptions, stateStorage, sp, changes, os.Stdout); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// if dry run, print the hint
		if dryRunParam {
			fmt.Printf("NOTE: Currently running in the --dry-run mode, the above configuration does not really take effect")
			render.Render(w, r, handler.SuccessResponse(ctx, "NOTE: Currently running in the --dry-run mode, the above configuration does not really take effect"))
			return
		}

		// Update LastSyncTimestamp to current time and set stack syncState to synced
		stackEntity.LastSyncTimestamp = time.Now()
		stackEntity.SyncState = constant.StackStateSynced

		// Update stack with repository
		err = h.stackRepo.Update(ctx, stackEntity)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Destroy completed
		logger.Info("apply completed")
		render.Render(w, r, handler.SuccessResponse(ctx, "apply completed"))

		// TODO: How to implement watch?
		// if o.Watch {
		// 	fmt.Println("Start watching changes ...")
		// 	if err = Watch(o, sp, changes); err != nil {
		// 		return err
		// 	}
		// }
	}
}

// @Summary      Destroy stack
// @Description  Destroy stack information by stack ID
// @Produce      json
// @Param        id   path      int                 true  "Stack ID"
// @Success      200  {object}  entity.Stack       "Success"
// @Failure      400  {object}  errors.DetailError  "Bad Request"
// @Failure      401  {object}  errors.DetailError  "Unauthorized"
// @Failure      429  {object}  errors.DetailError  "Too Many Requests"
// @Failure      404  {object}  errors.DetailError  "Not Found"
// @Failure      500  {object}  errors.DetailError  "Internal Server Error"
// @Router       /api/v1/stack/{stackID}/destroy [post]
func (h *Handler) DestroyStack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Getting stuff from context
		ctx := r.Context()
		logger := util.GetLogger(ctx)
		logger.Info("Destroying stack...")
		// Get params from URL parameter
		stackID := chi.URLParam(r, "stackID")
		// TODO: Define default behaviors
		// kpmParam, _ := strconv.ParseBool(r.URL.Query().Get("kpm"))
		// TODO: Should match automatically eventually
		workspaceParam := r.URL.Query().Get("workspace")
		detailParam, _ := strconv.ParseBool(r.URL.Query().Get("detail"))
		dryRunParam, _ := strconv.ParseBool(r.URL.Query().Get("dryrun"))

		// Get stack with repository
		id, err := strconv.Atoi(stackID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, ErrInvalidStacktID))
			return
		}
		stackEntity, err := h.stackRepo.Get(ctx, uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				render.Render(w, r, handler.FailureResponse(ctx, ErrGettingNonExistingStack))
				return
			}
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Get project by id
		project, err := stackEntity.Project.ConvertToCore()
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Get stack by id
		stack, err := stackEntity.ConvertToCore()
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// get workspace configurations
		// localBackend, err := backend.NewBackend("")
		// if err != nil {
		// 	render.Render(w, r, handler.FailureResponse(ctx, err))
		// 	return
		// }
		// wsStorage, err := bk.WorkspaceStorage()
		// if err != nil {
		// 	render.Render(w, r, handler.FailureResponse(ctx, err))
		// 	return
		// }
		// ws, err := wsStorage.Get(workspaceParam)
		// if err != nil {
		// 	render.Render(w, r, handler.FailureResponse(ctx, err))
		// 	return
		// }

		remoteBackend, err := h.GetBackendFromWorkspaceName(ctx, workspaceParam)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Build API inputs
		// get project to get source and workdir
		projectEntity, err := handler.GetProjectByID(ctx, h.projectRepo, stackEntity.Project.ID)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		directory, workDir, err := getWorkDirFromSource(ctx, stackEntity, projectEntity)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}
		destroyOptions := buildOptions(dryRunParam)
		stack.Path = workDir

		// Cleanup
		defer sourceapi.Cleanup(ctx, directory)

		// Compute state storage
		// TODO: this local storage is temporary, will support remote later
		stateStorage := remoteBackend.StateStorage(project.Name, stack.Name, workspaceParam)
		// logger.Info("Local state storage found", "Path", stateStorage)
		logger.Info("Remote state storage found", "Remote", stateStorage)

		priorState, err := stateStorage.Get()
		if err != nil || priorState == nil {
			logger.Info("can't find state", "project", project.Name, "stack", stack.Name, "workspace", workspaceParam)
			render.Render(w, r, handler.FailureResponse(ctx, ErrGettingNonExistingStateForStack))
			return
		}
		destroyResources := priorState.Resources

		if destroyResources == nil || len(priorState.Resources) == 0 {
			render.Render(w, r, handler.SuccessResponse(ctx, "No managed resources to destroy"))
			return
		}

		// compute changes for preview
		i := &apiv1.Spec{Resources: destroyResources}
		changes, err := engineapi.DestroyPreview(destroyOptions, i, project, stack, stateStorage)
		if err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Summary preview table
		changes.Summary(w, true)
		// detail detection
		if detailParam {
			changes.OutputDiff("all")
		}

		// if dryrun, print the hint
		if dryRunParam {
			fmt.Printf("Dry-run mode enabled, the above resources will be destroyed if dryrun is set to false")
			render.Render(w, r, handler.SuccessResponse(ctx, "Dry-run mode enabled, the above resources will be destroyed if dryrun is set to false"))
			return
		}

		// Destroy
		logger.Info("Start destroying resources......")
		if err = engineapi.Destroy(destroyOptions, i, changes, stateStorage); err != nil {
			render.Render(w, r, handler.FailureResponse(ctx, err))
			return
		}

		// Destroy completed
		logger.Info("destroy completed")
		render.Render(w, r, handler.SuccessResponse(ctx, "destroy completed"))

		// Cleanup
		sourceapi.Cleanup(ctx, directory)
	}
}

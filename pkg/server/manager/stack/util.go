package stack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"gorm.io/gorm"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/backend/storages"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	engineapi "kusionstack.io/kusion/pkg/engine/api"
	sourceapi "kusionstack.io/kusion/pkg/engine/api/source"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/state"
	"kusionstack.io/kusion/pkg/server/handler"
	"kusionstack.io/kusion/pkg/server/util"
)

func BuildOptions(dryrun bool) *engineapi.APIOptions {
	executeOptions := &engineapi.APIOptions{
		// Operator:     "operator",
		// Cluster:      "cluster",
		// IgnoreFields: []string{},
		DryRun: dryrun,
	}
	return executeOptions
}

// getWorkDirFromSource returns the workdir based on the source
// if the source type is local, it will return the path as an absolute path on the local filesystem
// if the source type is remote (git for example), it will pull the source and return the path to the pulled source
func GetWorkDirFromSource(ctx context.Context, stack *entity.Stack, project *entity.Project) (string, string, error) {
	logger := util.GetLogger(ctx)
	logger.Info("Getting workdir from stack source...")
	// TODO: Also copy the local workdir to /tmp directory?
	var err error
	directory := ""
	workDir := stack.Path

	if project.Source != nil && project.Source.SourceProvider != constant.SourceProviderTypeLocal {
		logger.Info("Non-local source provider, locating pulled source directory")
		// pull the latest source code
		directory, err = sourceapi.Pull(ctx, project.Source)
		if err != nil {
			return "", "", err
		}
		logger.Info("config pulled from source successfully", "directory", directory)
		workDir = filepath.Join(directory, stack.Path)
	}
	return directory, workDir, nil
}

func NewBackendFromEntity(backendEntity entity.Backend) (backend.Backend, error) {
	// TODO: refactor this so backend.NewBackend() share the same common logic
	var storage backend.Backend
	var err error
	switch backendEntity.BackendConfig.Type {
	case v1.BackendTypeLocal:
		bkConfig := backendEntity.BackendConfig.ToLocalBackend()
		if err = storages.CompleteLocalConfig(bkConfig); err != nil {
			return nil, fmt.Errorf("complete local config failed, %w", err)
		}
		return storages.NewLocalStorage(bkConfig), nil
	case v1.BackendTypeOss:
		bkConfig := backendEntity.BackendConfig.ToOssBackend()
		storages.CompleteOssConfig(bkConfig)
		if err = storages.ValidateOssConfig(bkConfig); err != nil {
			return nil, fmt.Errorf("invalid config of backend %s, %w", backendEntity.Name, err)
		}
		storage, err = storages.NewOssStorage(bkConfig)
		if err != nil {
			return nil, fmt.Errorf("new oss storage of backend %s failed, %w", backendEntity.Name, err)
		}
	case v1.BackendTypeS3:
		bkConfig := backendEntity.BackendConfig.ToS3Backend()
		storages.CompleteS3Config(bkConfig)
		if err = storages.ValidateS3Config(bkConfig); err != nil {
			return nil, fmt.Errorf("invalid config of backend %s: %w", backendEntity.Name, err)
		}
		storage, err = storages.NewS3Storage(bkConfig)
		if err != nil {
			return nil, fmt.Errorf("new s3 storage of backend %s failed, %w", backendEntity.Name, err)
		}
	default:
		return nil, fmt.Errorf("invalid type %s of backend %s", backendEntity.BackendConfig.Type, backendEntity.Name)
	}
	return storage, nil
}

func ProcessChanges(ctx context.Context, w http.ResponseWriter, changes *models.Changes, format string, detail bool) (string, error) {
	logger := util.GetLogger(ctx)
	logger.Info("Starting previewing stack in StackManager ...")

	if format == engineapi.JSONOutput {
		previewChanges, err := json.Marshal(changes)
		if err != nil {
			return "", err
		}
		logger.Info(string(previewChanges))
		return string(previewChanges), nil
	}

	if changes.AllUnChange() {
		logger.Info(NoDiffFound)
		return NoDiffFound, nil
	}

	// Summary preview table
	changes.Summary(w, true)
	// detail detection
	if detail {
		return changes.Diffs(true), nil
	}
	return "", nil
}

func (m *StackManager) getBackendFromWorkspaceName(ctx context.Context, workspaceName string) (backend.Backend, error) {
	logger := util.GetLogger(ctx)
	logger.Info("Getting backend based on workspace name...")
	// Get backend by id
	workspaceEntity, err := m.workspaceRepo.GetByName(ctx, workspaceName)
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil, err
	} else if err != nil {
		return nil, err
	}
	// Generate backend from entity
	remoteBackend, err := NewBackendFromEntity(*workspaceEntity.Backend)
	if err != nil {
		return nil, err
	}
	return remoteBackend, nil
}

func (m *StackManager) previewHelper(
	ctx context.Context,
	id uint,
	workspaceName string,
) (*v1.Spec, *models.Changes, state.Storage, error) {
	logger := util.GetLogger(ctx)
	logger.Info("Starting previewing stack in StackManager ...")

	// Get the stack entity by id
	stackEntity, err := m.stackRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, ErrGettingNonExistingStack
		}
		return nil, nil, nil, err
	}

	// Get project by id
	project, err := stackEntity.Project.ConvertToCore()
	if err != nil {
		return nil, nil, nil, err
	}

	// Get stack by id
	stack, err := stackEntity.ConvertToCore()
	if err != nil {
		return nil, nil, nil, err
	}

	// Get backend from workspace name
	stackBackend, err := m.getBackendFromWorkspaceName(ctx, workspaceName)
	if err != nil {
		return nil, nil, nil, err
	}

	// Get workspace configurations from backend
	// TODO: temporarily local for now, should be replaced by variable sets
	wsStorage, err := stackBackend.WorkspaceStorage()
	if err != nil {
		return nil, nil, nil, err
	}
	ws, err := wsStorage.Get(workspaceName)
	if err != nil {
		return nil, nil, nil, err
	}
	// Compute state storage
	stateStorage := stackBackend.StateStorage(project.Name, ws.Name)
	logger.Info("Local state storage found", "Path", stateStorage)

	// Build API inputs
	// get project to get source and workdir
	projectEntity, err := handler.GetProjectByID(ctx, m.projectRepo, stackEntity.Project.ID)
	if err != nil {
		return nil, nil, nil, err
	}

	directory, workDir, err := GetWorkDirFromSource(ctx, stackEntity, projectEntity)
	if err != nil {
		return nil, nil, nil, err
	}
	executeOptions := BuildOptions(false)
	stack.Path = workDir

	// Cleanup
	defer sourceapi.Cleanup(ctx, directory)

	// Generate spec
	sp, err := engineapi.GenerateSpecWithSpinner(project, stack, ws, true)
	if err != nil {
		return nil, nil, nil, err
	}

	// return immediately if no resource found in stack
	// todo: if there is no resource, should still do diff job; for now, if output is json format, there is no hint
	if sp == nil || len(sp.Resources) == 0 {
		logger.Info("No resource change found in this stack...")
		return nil, nil, nil, nil
	}

	changes, err := engineapi.Preview(executeOptions, stateStorage, sp, project, stack)
	return sp, changes, stateStorage, err
}

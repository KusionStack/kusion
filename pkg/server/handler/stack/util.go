package stack

import (
	"context"
	"fmt"
	"path/filepath"

	"gorm.io/gorm"
	v1 "kusionstack.io/kusion/pkg/apis/internal.kusion.io/v1"
	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/backend/storages"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	engineapi "kusionstack.io/kusion/pkg/engine/api"
	sourceapi "kusionstack.io/kusion/pkg/engine/api/source"
	"kusionstack.io/kusion/pkg/server/util"
)

func buildOptions(dryrun bool) *engineapi.APIOptions {
	// Construct intent options
	// intentOptions := &buildersapi.Options{
	// 	IsKclPkg:  kpmParam,
	// 	WorkDir:   workDir,
	// 	Arguments: map[string]string{},
	// 	NoStyle:   true,
	// }
	// Construct preview api option
	// TODO: Complete preview options
	// TODO: Operator should be derived from auth info
	// TODO: Cluster should be derived from workspace config
	previewOptions := &engineapi.APIOptions{
		// Operator:     "operator",
		// Cluster:      "cluster",
		// IgnoreFields: []string{},
		DryRun: dryrun,
	}
	// return intentOptions, previewOptions
	return previewOptions
}

// getWorkDirFromSource returns the workdir based on the source
// if the source type is local, it will return the path as an absolute path on the local filesystem
// if the source type is remote (git for example), it will pull the source and return the path to the pulled source
func getWorkDirFromSource(ctx context.Context, stack *entity.Stack, project *entity.Project) (string, string, error) {
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
	// var emptyCfg bool
	// cfg, err := config.GetConfig()
	// if errors.Is(err, config.ErrEmptyConfig) {
	// 	emptyCfg = true
	// } else if err != nil {
	// 	return nil, err
	// } else if cfg.Backends == nil {
	// 	emptyCfg = true
	// }

	// var bkCfg *v1.BackendConfig
	// if name == "" && (emptyCfg || cfg.Backends.Current == "") {
	// 	// if empty backends config or empty current backend, use default local storage
	// 	bkCfg = &v1.BackendConfig{Type: v1.BackendTypeLocal}
	// } else {
	// 	if name == "" {
	// 		name = cfg.Backends.Current
	// 	}
	// 	bkCfg = cfg.Backends.Backends[name]
	// 	if bkCfg == nil {
	// 		return nil, fmt.Errorf("config of backend %s does not exist", name)
	// 	}
	// }

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
	case v1.BackendTypeMysql:
		bkConfig := backendEntity.BackendConfig.ToMysqlBackend()
		storages.CompleteMysqlConfig(bkConfig)
		if err = storages.ValidateMysqlConfig(bkConfig); err != nil {
			return nil, fmt.Errorf("invalid config of backend %s, %w", backendEntity.Name, err)
		}
		storage, err = storages.NewMysqlStorage(bkConfig)
		if err != nil {
			return nil, fmt.Errorf("new mysql storage of backend %s failed, %w", backendEntity.Name, err)
		}
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

func (h *Handler) GetBackendFromWorkspaceName(ctx context.Context, workspaceName string) (backend.Backend, error) {
	logger := util.GetLogger(ctx)
	logger.Info("Getting backend based on workspace name...")
	// Get backend by id
	workspaceEntity, err := h.workspaceRepo.GetByName(ctx, workspaceName)
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

package backend

import (
	"errors"
	"fmt"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/backend/storages"
	"kusionstack.io/kusion/pkg/config"
	"kusionstack.io/kusion/pkg/engine/spec"
	"kusionstack.io/kusion/pkg/engine/state"
	"kusionstack.io/kusion/pkg/workspace"
)

// Backend is used to provide the storage service for Workspace, Spec and State.
type Backend interface {
	// SpecStorage returns the spec storage.
	SpecStorage(project, stack, workspace string) spec.Storage

	// WorkspaceStorage returns the workspace storage and init default workspace.
	WorkspaceStorage() (workspace.Storage, error)

	// StateStorage returns the state storage.
	StateStorage(project, stack, workspace string) state.Storage
}

// NewBackend creates the Backend with the configuration set in the Kusion configuration file, where the input
// is the configured backend name. If the backend configuration is invalid, NewBackend will get failed. If the
// input name is empty, use the current backend. If no current backend is specified or backends config is empty,
// and the input name is empty, use the default local storage.
func NewBackend(name string) (Backend, error) {
	var emptyCfg bool
	cfg, err := config.GetConfig()
	if errors.Is(err, config.ErrEmptyConfig) {
		emptyCfg = true
	} else if err != nil {
		return nil, err
	} else if cfg.Backends == nil {
		emptyCfg = true
	}

	var bkCfg *v1.BackendConfig
	if name == "" && (emptyCfg || cfg.Backends.Current == "") {
		// if empty backends config or empty current backend, use default local storage
		bkCfg = &v1.BackendConfig{Type: v1.BackendTypeLocal}
	} else {
		if name == "" {
			name = cfg.Backends.Current
		}
		bkCfg = cfg.Backends.Backends[name]
		if bkCfg == nil {
			return nil, fmt.Errorf("config of backend %s does not exist", name)
		}
	}

	var storage Backend
	switch bkCfg.Type {
	case v1.BackendTypeLocal:
		bkConfig := bkCfg.ToLocalBackend()
		if err = storages.CompleteLocalConfig(bkConfig); err != nil {
			return nil, fmt.Errorf("complete local config failed, %w", err)
		}
		return storages.NewLocalStorage(bkConfig), nil
	case v1.BackendTypeMysql:
		bkConfig := bkCfg.ToMysqlBackend()
		storages.CompleteMysqlConfig(bkConfig)
		if err = storages.ValidateMysqlConfig(bkConfig); err != nil {
			return nil, fmt.Errorf("invalid config of backend %s, %w", name, err)
		}
		storage, err = storages.NewMysqlStorage(bkConfig)
		if err != nil {
			return nil, fmt.Errorf("new mysql storage of backend %s failed, %w", name, err)
		}
	case v1.BackendTypeOss:
		bkConfig := bkCfg.ToOssBackend()
		storages.CompleteOssConfig(bkConfig)
		if err = storages.ValidateOssConfig(bkConfig); err != nil {
			return nil, fmt.Errorf("invalid config of backend %s, %w", name, err)
		}
		storage, err = storages.NewOssStorage(bkConfig)
		if err != nil {
			return nil, fmt.Errorf("new oss storage of backend %s failed, %w", name, err)
		}
	case v1.BackendTypeS3:
		bkConfig := bkCfg.ToS3Backend()
		storages.CompleteS3Config(bkConfig)
		if err = storages.ValidateS3Config(bkConfig); err != nil {
			return nil, fmt.Errorf("invalid config of backend %s: %w", name, err)
		}
		storage, err = storages.NewS3Storage(bkConfig)
		if err != nil {
			return nil, fmt.Errorf("new s3 storage of backend %s failed, %w", name, err)
		}
	default:
		return nil, fmt.Errorf("invalid type %s of backend %s", bkCfg.Type, name)
	}
	return storage, nil
}

// NewWorkspaceStorage calls NewBackend and WorkspaceStorage to new a workspace storage from specified backend.
func NewWorkspaceStorage(backendName string) (workspace.Storage, error) {
	bk, err := NewBackend(backendName)
	if err != nil {
		return nil, err
	}
	return bk.WorkspaceStorage()
}

// NewSpecStorage calls NewBackend and returns a spec storage from specified backend.
func NewSpecStorage(backendName string, project, stack, workspace string) (spec.Storage, error) {
	bk, err := NewBackend(backendName)
	if err != nil {
		return nil, err
	}
	return bk.SpecStorage(project, stack, workspace), nil
}

package workspace

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"gorm.io/gorm"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/request"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"

	"github.com/elliotchance/orderedmap/v2"
	kpmdownloader "kcl-lang.io/kpm/pkg/downloader"
	kpmpkg "kcl-lang.io/kpm/pkg/package"
)

func (m *WorkspaceManager) GetWorkspaceConfigs(ctx context.Context, id uint) (*request.WorkspaceConfigs, error) {
	workspaceEntity, err := m.workspaceRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingWorkspace
		}
		return nil, err
	}

	// Get backend from workspace name.
	wsBackend, err := m.getBackendFromWorkspaceName(ctx, workspaceEntity.Name)
	if err != nil {
		return nil, err
	}

	// Get workspace storage from backend.
	wsStorage, err := wsBackend.WorkspaceStorage()
	if err != nil {
		return nil, err
	}

	// Get workspace configurations from storage.
	ws, err := wsStorage.Get(workspaceEntity.Name)
	if err != nil {
		return nil, err
	}

	return &request.WorkspaceConfigs{
		Workspace: ws,
	}, nil
}

func (m *WorkspaceManager) ValidateWorkspaceConfigs(ctx context.Context, configs request.WorkspaceConfigs) (*request.WorkspaceConfigs, error) {
	// Validate the workspace configs to be updated.
	if err := m.validateWorkspaceConfigs(ctx, configs.Workspace); err != nil {
		return nil, err
	}

	return &configs, nil
}

func (m *WorkspaceManager) UpdateWorkspaceConfigs(ctx context.Context, id uint, configs request.WorkspaceConfigs) (*request.WorkspaceConfigs, error) {
	workspaceEntity, err := m.workspaceRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingWorkspace
		}
		return nil, err
	}

	// Get backend from workspace name.
	wsBackend, err := m.getBackendFromWorkspaceName(ctx, workspaceEntity.Name)
	if err != nil {
		return nil, err
	}

	// Get workspace storage from backend.
	wsStorage, err := wsBackend.WorkspaceStorage()
	if err != nil {
		return nil, err
	}

	// Validate the workspace configs to be updated.
	if configs.Workspace.Name != "" && configs.Workspace.Name != workspaceEntity.Name {
		return nil, fmt.Errorf("inconsistent workspace name, want: %s, got: %s", workspaceEntity.Name, configs.Workspace.Name)
	} else {
		configs.Workspace.Name = workspaceEntity.Name
	}

	if err = m.validateWorkspaceConfigs(ctx, configs.Workspace); err != nil {
		return nil, err
	}

	// Update workspace configs in the storage.
	if err = wsStorage.Update(configs.Workspace); err != nil {
		return nil, err
	}

	return &configs, nil
}

func (m *WorkspaceManager) CreateKCLModDependencies(ctx context.Context, id uint) (string, error) {
	workspaceEntity, err := m.workspaceRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrGettingNonExistingWorkspace
		}
		return "", err
	}

	// Get backend from workspace name.
	wsBackend, err := m.getBackendFromWorkspaceName(ctx, workspaceEntity.Name)
	if err != nil {
		return "", err
	}

	// Get workspace storage from backend.
	wsStorage, err := wsBackend.WorkspaceStorage()
	if err != nil {
		return "", err
	}

	// Get workspace configurations from storage.
	ws, err := wsStorage.Get(workspaceEntity.Name)
	if err != nil {
		return "", err
	}

	// Generate the dependencies in `kcl.mod`.
	deps := &kpmpkg.Dependencies{
		Deps: orderedmap.NewOrderedMap[string, kpmpkg.Dependency](),
	}

	// Traverse the modules in the workspace.
	for modName, modConfig := range ws.Modules {
		// Parse the source url of the module.
		src, err := kpmdownloader.NewSourceFromStr(modConfig.Path)
		if err != nil {
			return "", err
		}

		// Prepare the dependency object.
		dep := kpmpkg.Dependency{
			Name:    modName,
			Version: modConfig.Version,
		}

		if src.Git != nil {
			dep.Source = kpmdownloader.Source{
				Git: &kpmdownloader.Git{
					Url: modConfig.Path,
					Tag: modConfig.Version,
				},
			}
		} else if src.Oci != nil {
			u, _ := url.Parse(modConfig.Path)
			dep.Source = kpmdownloader.Source{
				Oci: &kpmdownloader.Oci{
					Reg:  u.Host,
					Repo: strings.TrimPrefix(u.Path, "/"),
					Tag:  modConfig.Version,
				},
			}
		} else if src.Local != nil {
			dep.Source = kpmdownloader.Source{
				Local: src.Local,
			}
		}

		deps.Deps.Set(modName, dep)
	}

	return deps.MarshalTOML(), nil
}

func (m *WorkspaceManager) getBackendFromWorkspaceName(ctx context.Context, workspaceName string) (backend.Backend, error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Getting backend based on workspace name...")

	var remoteBackend backend.Backend
	if workspaceName == constant.DefaultWorkspace {
		// Get the default backend.
		return m.getDefaultBackend()
	} else {
		// Get workspace entity by name.
		workspaceEntity, err := m.workspaceRepo.GetByName(ctx, workspaceName)
		if err != nil && err == gorm.ErrRecordNotFound {
			return nil, ErrGettingNonExistingWorkspace
		} else if err != nil {
			return nil, err
		}

		// Generate backend from the workspace entity.
		remoteBackend, err = NewBackendFromEntity(*workspaceEntity.Backend)
		if err != nil {
			return nil, err
		}
	}

	return remoteBackend, nil
}

func (m *WorkspaceManager) getDefaultBackend() (backend.Backend, error) {
	defaultBackendEntity := m.defaultBackend
	remoteBackend, err := NewBackendFromEntity(defaultBackendEntity)
	if err != nil {
		return nil, err
	}

	return remoteBackend, nil
}

func (m *WorkspaceManager) validateWorkspaceConfigs(ctx context.Context, workspaceConfigs *v1.Workspace) error {
	logger := logutil.GetLogger(ctx)
	logger.Info("Validating workspace configs...")

	var modulesNotFound, modulesPathNotMatched []string
	for moduleName, moduleConfigs := range workspaceConfigs.Modules {
		// Get module entity by name.
		moduleEntity, err := m.moduleRepo.Get(ctx, moduleName)
		if err != nil {
			// The modules declared in the workspace should be registered.
			if errors.Is(err, gorm.ErrRecordNotFound) {
				modulesNotFound = append(modulesNotFound, moduleName)
			} else {
				return err
			}
		} else {
			// The oci path of the modules should match the registered information.
			if moduleConfigs.Path != "" && moduleConfigs.Path != moduleEntity.URL.String() {
				modulesPathNotMatched = append(modulesPathNotMatched, moduleName)
			} else if moduleConfigs.Path == "" {
				// Set the oci path with the registered information.
				moduleConfigs.Path = moduleEntity.URL.String()
			}
		}
	}

	// Prepare and return the errors according to the results.
	errModulesNotFound := fmt.Errorf(ErrMsgModulesNotRegistered,
		plural(len(modulesNotFound)), modulesNotFound, verb(len(modulesNotFound)))
	errModulesPathNotMatched := fmt.Errorf(ErrMsgModulesPathNotMatched,
		plural(len(modulesPathNotMatched)), modulesPathNotMatched, verb(len(modulesPathNotMatched)))

	if len(modulesNotFound) > 0 && len(modulesPathNotMatched) > 0 {
		return errors.Join(errModulesNotFound, errModulesPathNotMatched)
	} else if len(modulesNotFound) > 0 {
		return errModulesNotFound
	} else if len(modulesPathNotMatched) > 0 {
		return errModulesPathNotMatched
	}

	return nil
}

package workspace

import (
	"fmt"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/backend/storages"
	"kusionstack.io/kusion/pkg/domain/entity"
)

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

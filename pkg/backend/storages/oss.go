package storages

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/engine/release"
	releasestorages "kusionstack.io/kusion/pkg/engine/release/storages"
	"kusionstack.io/kusion/pkg/workspace"
	workspacestorages "kusionstack.io/kusion/pkg/workspace/storages"
)

// OssStorage is an implementation of backend.Backend which uses oss as storage.
type OssStorage struct {
	bucket *oss.Bucket

	// prefix will be added to the object storage key, so that all the files are stored under the prefix.
	prefix string
}

func NewOssStorage(config *v1.BackendOssConfig) (*OssStorage, error) {
	client, err := oss.New(config.Endpoint, config.AccessKeyID, config.AccessKeySecret)
	if err != nil {
		return nil, err
	}
	bucket, err := client.Bucket(config.Bucket)
	if err != nil {
		return nil, err
	}

	return &OssStorage{bucket: bucket, prefix: config.Prefix}, nil
}

func (s *OssStorage) WorkspaceStorage() (workspace.Storage, error) {
	return workspacestorages.NewOssStorage(s.bucket, workspacestorages.GenGenericOssWorkspacePrefixKey(s.prefix))
}

func (s *OssStorage) ReleaseStorage(project, workspace string) (release.Storage, error) {
	return releasestorages.NewOssStorage(s.bucket, releasestorages.GenGenericOssReleasePrefixKey(s.prefix, project, workspace))
}

func (s *OssStorage) StateStorageWithPath(path string) (release.Storage, error) {
	return releasestorages.NewOssStorage(s.bucket, releasestorages.GenReleasePrefixKeyWithPath(s.prefix, path))
}

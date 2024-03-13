package storages

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/engine/state"
	statestorages "kusionstack.io/kusion/pkg/engine/state/storages"
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

func (s *OssStorage) StateStorage(project, stack, workspace string) state.Storage {
	return statestorages.NewOssStorage(s.bucket, statestorages.GenGenericOssStateFileKey(s.prefix, project, stack, workspace))
}

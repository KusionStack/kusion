package storages

import (
	"context"

	google "cloud.google.com/go/storage"
	"google.golang.org/api/option"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/engine/release"
	releasestorages "kusionstack.io/kusion/pkg/engine/release/storages"
	"kusionstack.io/kusion/pkg/engine/resource/graph"
	graphstorages "kusionstack.io/kusion/pkg/engine/resource/graph/storages"
	projectstorages "kusionstack.io/kusion/pkg/project/storages"
	"kusionstack.io/kusion/pkg/workspace"
	workspacestorages "kusionstack.io/kusion/pkg/workspace/storages"
)

// GoogleStorage is an implementation of backend.Backend which uses google cloud as storage.
type GoogleStorage struct {
	bucket *google.BucketHandle

	// prefix will be added to the object storage key, so that all the files are stored under the prefix.
	prefix string
}

func NewGoogleStorage(config *v1.BackendGoogleConfig) (*GoogleStorage, error) {
	client, err := google.NewClient(context.Background(), option.WithCredentials(config.Credentials))
	if err != nil {
		return nil, err
	}
	bucket := client.Bucket(config.Bucket)

	return &GoogleStorage{
		bucket: bucket,
		prefix: config.Prefix,
	}, nil
}

func (s *GoogleStorage) WorkspaceStorage() (workspace.Storage, error) {
	return workspacestorages.NewGoogleStorage(s.bucket, workspacestorages.GenGenericOssWorkspacePrefixKey(s.prefix))
}

func (s *GoogleStorage) ReleaseStorage(project, workspace string) (release.Storage, error) {
	return releasestorages.NewGoogleStorage(s.bucket, releasestorages.GenGenericOssReleasePrefixKey(s.prefix, project, workspace))
}

func (s *GoogleStorage) StateStorageWithPath(path string) (release.Storage, error) {
	return releasestorages.NewGoogleStorage(s.bucket, releasestorages.GenReleasePrefixKeyWithPath(s.prefix, path))
}

func (s *GoogleStorage) GraphStorage(project, workspace string) (graph.Storage, error) {
	return graphstorages.NewGoogleStorage(s.bucket, graphstorages.GenGenericOssResourcePrefixKey(s.prefix, project, workspace))
}

func (s *GoogleStorage) ProjectStorage() (map[string][]string, error) {
	return projectstorages.NewGoogleStorage(s.bucket, projectstorages.GenGenericOssReleasePrefixKey(s.prefix)).Get()
}

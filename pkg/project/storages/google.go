package storages

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/iterator"

	googlestorage "cloud.google.com/go/storage"
)

// GoogleStorage is an implementation of graph.Storage which uses google cloud as storage.
type GoogleStorage struct {
	bucket googlestorage.BucketHandle
	prefix string
}

// NewGoogleStorage creates a new GoogleStorage instance.
func NewGoogleStorage(bucket *googlestorage.BucketHandle, prefix string) *GoogleStorage {
	s := &GoogleStorage{
		bucket: *bucket,
		prefix: prefix,
	}
	return s
}

// Get returns a project map which key is workspace name and value is its belonged project list.
func (s *GoogleStorage) Get() (map[string][]string, error) {
	ctx := context.Background()
	projects := map[string][]string{}
	projectQuery := &googlestorage.Query{
		Prefix:    s.prefix + "/",
		Delimiter: "/",
	}
	it := s.bucket.Objects(ctx, projectQuery)
	for {
		project, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list projects directory from google storage failed: %w", err)
		}
		projectDir := strings.TrimPrefix(project.Name, s.prefix+"/")
		projectDir = strings.TrimSuffix(projectDir, "/")

		// List workspaces under the project prefix
		wsQuery := &googlestorage.Query{
			Prefix:    project.Prefix + "/",
			Delimiter: "/",
		}
		wsIt := s.bucket.Objects(ctx, wsQuery)
		for {
			workspace, err := wsIt.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("list project's workspaces directory from google storage failed: %w", err)
			}
			// Get each of the workspace name
			workspaceDir := strings.TrimPrefix(workspace.Prefix, project.Prefix)
			workspaceDir = strings.TrimSuffix(workspaceDir, "/")
			// Store workspace name as key, project name as value
			projects[workspaceDir] = append(projects[workspaceDir], projectDir)
		}
	}
	return projects, nil
}

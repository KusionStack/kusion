package storages

import (
	"fmt"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OssStorage is an implementation of project.Storage which uses oss as storage.
type OssStorage struct {
	bucket *oss.Bucket

	// The prefix to store the project folders' directory.
	prefix string
}

// NewOssStorage creates a new OssStorage instance.
func NewOssStorage(bucket *oss.Bucket, prefix string) *OssStorage {
	s := &OssStorage{
		bucket: bucket,
		prefix: prefix,
	}

	return s
}

// Get returns a project map which key is workspace name and value is its belonged project list.
func (s *OssStorage) Get() (map[string][]string, error) {
	projects := map[string][]string{}
	marker := oss.Marker("")
	for {
		// list all the objects under the release prefix
		projectObj, err := s.bucket.ListObjects(oss.Prefix(s.prefix+"/"), marker, oss.Delimiter("/"))
		if err != nil {
			return nil, fmt.Errorf("list projects directory from oss failed: %w", err)
		}

		for _, projectPrefix := range projectObj.CommonPrefixes {
			// Get project name
			projectDir := strings.TrimPrefix(projectPrefix, s.prefix+"/")
			projectDir = strings.TrimSuffix(projectDir, "/")

			// List workspaces under the project prefix
			workspaces, err := s.bucket.ListObjects(oss.Prefix(projectPrefix), oss.Delimiter("/"))
			if err != nil {
				return nil, fmt.Errorf("list project's workspaces directory from oss failed: %w", err)
			}

			for _, workspacePrefix := range workspaces.CommonPrefixes {
				workspaceDir := strings.TrimPrefix(workspacePrefix, projectPrefix)
				workspaceDir = strings.TrimSuffix(workspaceDir, "/")

				// Store workspace name as key, project name as value
				projects[workspaceDir] = append(projects[workspaceDir], projectDir)
			}
		}

		// Break if there are no more results
		if projectObj.IsTruncated {
			marker = oss.Marker(projectObj.NextMarker)
		} else {
			break
		}
	}
	return projects, nil
}

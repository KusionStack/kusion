package storages

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

const (
	resourcesPrefix = "resources"
	graphFileName   = "graph.json"
)

var (
	ErrGraphNotExist     = errors.New("graph does not exist")
	ErrGraphAlreadyExist = errors.New("graph has already existed")
)

// GenResourceDirPath generates the resource dir path, which is used for LocalStorage.
func GenGraphDirPath(dir, project, workspace string) string {
	return filepath.Join(dir, resourcesPrefix, project, workspace)
}

// GenGenericOssResourcePrefixKey generates generic oss resource prefix, which is use for OssStorage and S3Storage.
func GenGenericOssResourcePrefixKey(prefix, project, workspace string) string {
	prefix = strings.TrimPrefix(prefix, "/")
	if prefix != "" {
		prefix += "/"
	}

	return fmt.Sprintf("%s%s/%s/%s", prefix, resourcesPrefix, project, workspace)
}

// GenResourcePrefixKeyWithPath generates oss state file key with cloud and env instead of workspace, which is use for OssStorage and S3Storage.
func GenResourcePrefixKeyWithPath(prefix, path string) string {
	prefix = strings.TrimPrefix(prefix, "/")
	if prefix != "" {
		prefix += "/"
	}

	return fmt.Sprintf("%s%s/%s", prefix, resourcesPrefix, path)
}

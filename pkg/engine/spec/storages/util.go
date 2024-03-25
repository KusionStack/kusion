package storages

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	specFile    = "spec.yaml"
	specsPrefix = "specs"
)

// GetSpecFilePath returns the location on the disk where the Spec data should be present
func GetSpecFilePath(dir, project, stack, workspace string) string {
	return filepath.Join(dir, specsPrefix, project, stack, workspace, specFile)
}

// GetObjectStoreSpecFileKey returns the object storage key, which is use for OssStorage and S3Storage.
func GetObjectStoreSpecFileKey(prefix, project, stack, workspace string) string {
	prefix = strings.TrimPrefix(prefix, "/")
	if prefix != "" {
		prefix += "/"
	}
	return fmt.Sprintf("%s%s/%s/%s/%s/%s", prefix, specsPrefix, project, stack, workspace, specFile)
}

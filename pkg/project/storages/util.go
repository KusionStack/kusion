package storages

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	releasesPrefix = "releases"
)

// GenProjectDirPath returns the project dir path, which is used for LocalStorage.
func GenProjectDirPath(dir string) string {
	return filepath.Join(dir, releasesPrefix)
}

// GenGenericOssReleasePrefixKey generates generic oss release prefix, which is use for OssStorage and S3Storage.
func GenGenericOssReleasePrefixKey(prefix string) string {
	prefix = strings.TrimPrefix(prefix, "/")
	if prefix != "" {
		prefix += "/"
	}
	return fmt.Sprintf("%s%s", prefix, releasesPrefix)
}

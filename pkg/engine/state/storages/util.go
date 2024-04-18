package storages

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	stateFile    = "state.yaml"
	stateTable   = "state"
	statesPrefix = "states"
)

// GenStateFilePath generates the state file path, which is used for LocalStorage.
func GenStateFilePath(dir, project, workspace string) string {
	return filepath.Join(dir, statesPrefix, project, workspace, stateFile)
}

// GenGenericOssStateFileKey generates generic oss state file key, which is use for OssStorage and S3Storage.
func GenGenericOssStateFileKey(prefix, project, workspace string) string {
	prefix = strings.TrimPrefix(prefix, "/")
	if prefix != "" {
		prefix += "/"
	}
	return fmt.Sprintf("%s%s/%s/%s/%s", prefix, statesPrefix, project, workspace, stateFile)
}

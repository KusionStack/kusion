package storages

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

const (
	releasesPrefix = "releases"
	metadataFile   = ".metadata.yml"
	yamlSuffix     = ".yaml"
)

var (
	ErrReleaseNotExist     = errors.New("release does not exist")
	ErrReleaseAlreadyExist = errors.New("release has already existed")
)

// GenReleaseDirPath generates the release dir path, which is used for LocalStorage.
func GenReleaseDirPath(dir, project, workspace string) string {
	return filepath.Join(dir, releasesPrefix, project, workspace)
}

// GenGenericOssReleasePrefixKey generates generic oss release prefix, which is use for OssStorage and S3Storage.
func GenGenericOssReleasePrefixKey(prefix, project, workspace string) string {
	prefix = strings.TrimPrefix(prefix, "/")
	if prefix != "" {
		prefix += "/"
	}
	return fmt.Sprintf("%s%s/%s/%s", prefix, releasesPrefix, project, workspace)
}

// GenReleasePrefixKeyWithPath generates oss state file key with cloud and env instead of workspace, which is use for OssStorage and S3Storage.
func GenReleasePrefixKeyWithPath(prefix, path string) string {
	prefix = strings.TrimPrefix(prefix, "/")
	if prefix != "" {
		prefix += "/"
	}
	return fmt.Sprintf("%s%s/%s", prefix, releasesPrefix, path)
}

// releasesMetaData contains mata data of the releases of a specified project and workspace. The mata data
// includes the latest revision, and synopsis of the releases.
type releasesMetaData struct {
	// LatestRevision of the Releases.
	LatestRevision uint64 `yaml:"latestRevision,omitempty" json:"latestRevision,omitempty"`

	// ReleaseMetaDatas are the mata data of the Releases.
	ReleaseMetaDatas []*releaseMetaData `yaml:"releaseMetaDatas,omitempty" json:"releaseMetaDatas,omitempty"`
}

// releaseMetaData contains mata data of a specified release, which contains the Revision and Stack.
type releaseMetaData struct {
	// Revision of the Release.
	Revision uint64

	// Stack of the Release.
	Stack string
}

// checkRevisionExistence returns the workspace exists or not.
func checkRevisionExistence(meta *releasesMetaData, revision uint64) bool {
	for _, metaData := range meta.ReleaseMetaDatas {
		if revision == metaData.Revision {
			return true
		}
	}
	return false
}

// getRevisions returns all the release revisions of a project and workspace.
func getRevisions(meta *releasesMetaData) []uint64 {
	var revisions []uint64
	for _, release := range meta.ReleaseMetaDatas {
		if release != nil {
			revisions = append(revisions, release.Revision)
		}
	}
	return revisions
}

// getStackBoundRevisions returns the release revisions of a project, workspace and stack.
func getStackBoundRevisions(meta *releasesMetaData, stack string) []uint64 {
	var revisions []uint64
	for _, release := range meta.ReleaseMetaDatas {
		if release != nil && release.Stack == stack {
			revisions = append(revisions, release.Revision)
		}
	}
	return revisions
}

// addLatestReleaseMetaData adds a release and updates the latest revision in the metadata, called
// by the storage.Create.
func addLatestReleaseMetaData(meta *releasesMetaData, revision uint64, stack string) {
	meta.LatestRevision = revision
	metaData := &releaseMetaData{
		Revision: revision,
		Stack:    stack,
	}
	meta.ReleaseMetaDatas = append(meta.ReleaseMetaDatas, metaData)
}

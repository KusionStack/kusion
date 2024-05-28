package storages

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

// LocalStorage is an implementation of release.Storage which uses local filesystem as storage.
type LocalStorage struct {
	// The directory path to store the release files.
	path string

	meta *releasesMetaData
}

// NewLocalStorage news local release storage, and derives metadata.
func NewLocalStorage(path string) (*LocalStorage, error) {
	s := &LocalStorage{path: path}

	// create the releases directory
	if err := os.MkdirAll(s.path, os.ModePerm); err != nil {
		return nil, fmt.Errorf("create releases directory failed, %w", err)
	}
	// read releases metadata
	if err := s.readMeta(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *LocalStorage) Get(revision uint64) (*v1.Release, error) {
	if !checkRevisionExistence(s.meta, revision) {
		return nil, ErrReleaseNotExist
	}

	content, err := os.ReadFile(filepath.Join(s.path, fmt.Sprintf("%d%s", revision, yamlSuffix)))
	if err != nil {
		return nil, fmt.Errorf("read release file failed: %w", err)
	}

	r := &v1.Release{}
	if err = yaml.Unmarshal(content, r); err != nil {
		return nil, fmt.Errorf("yaml unmarshal release failed: %w", err)
	}
	return r, nil
}

func (s *LocalStorage) GetRevisions() []uint64 {
	return getRevisions(s.meta)
}

func (s *LocalStorage) GetStackBoundRevisions(stack string) []uint64 {
	return getStackBoundRevisions(s.meta, stack)
}

func (s *LocalStorage) GetLatestRevision() uint64 {
	return s.meta.LatestRevision
}

func (s *LocalStorage) Create(r *v1.Release) error {
	if checkRevisionExistence(s.meta, r.Revision) {
		return ErrReleaseAlreadyExist
	}

	if err := s.writeRelease(r); err != nil {
		return err
	}

	addLatestReleaseMetaData(s.meta, r.Revision, r.Stack)
	return s.writeMeta()
}

func (s *LocalStorage) Update(r *v1.Release) error {
	if !checkRevisionExistence(s.meta, r.Revision) {
		return ErrReleaseNotExist
	}

	return s.writeRelease(r)
}

func (s *LocalStorage) readMeta() error {
	content, err := os.ReadFile(filepath.Join(s.path, metadataFile))
	if os.IsNotExist(err) {
		s.meta = &releasesMetaData{}
		return nil
	} else if err != nil {
		return fmt.Errorf("read releases metadata file failed: %w", err)
	}

	meta := &releasesMetaData{}
	if err = yaml.Unmarshal(content, meta); err != nil {
		return fmt.Errorf("yaml unmarshal releases metadata failed: %w", err)
	}
	s.meta = meta
	return nil
}

func (s *LocalStorage) writeMeta() error {
	content, err := yaml.Marshal(s.meta)
	if err != nil {
		return fmt.Errorf("yaml marshal releases metadata failed: %w", err)
	}

	if err = os.WriteFile(filepath.Join(s.path, metadataFile), content, os.ModePerm); err != nil {
		return fmt.Errorf("write releases metadata file failed: %w", err)
	}
	return nil
}

func (s *LocalStorage) writeRelease(r *v1.Release) error {
	content, err := yaml.Marshal(r)
	if err != nil {
		return fmt.Errorf("yaml marshal release failed: %w", err)
	}

	if err = os.WriteFile(filepath.Join(s.path, fmt.Sprintf("%d%s", r.Revision, yamlSuffix)), content, os.ModePerm); err != nil {
		return fmt.Errorf("write release file failed: %w", err)
	}
	return nil
}

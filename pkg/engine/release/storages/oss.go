package storages

import (
	"bytes"
	"fmt"
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

// OssStorage is an implementation of release.Storage which uses oss as storage.
type OssStorage struct {
	bucket *oss.Bucket

	// The prefix to store the release files.
	prefix string

	meta *releasesMetaData
}

// NewOssStorage news oss release storage, and derives metadata.
func NewOssStorage(bucket *oss.Bucket, prefix string) (*OssStorage, error) {
	s := &OssStorage{
		bucket: bucket,
		prefix: prefix,
	}
	if err := s.readMeta(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *OssStorage) Get(revision uint64) (*v1.Release, error) {
	if !checkRevisionExistence(s.meta, revision) {
		return nil, ErrReleaseNotExist
	}

	body, err := s.bucket.GetObject(fmt.Sprintf("%s/%d%s", s.prefix, revision, yamlSuffix))
	if err != nil {
		return nil, fmt.Errorf("get release from oss failed: %w", err)
	}
	defer func() {
		_ = body.Close()
	}()
	content, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("read release failed: %w", err)
	}

	r := &v1.Release{}
	if err = yaml.Unmarshal(content, r); err != nil {
		return nil, fmt.Errorf("yaml unmarshal release failed: %w", err)
	}
	return r, nil
}

func (s *OssStorage) GetRevisions() []uint64 {
	return getRevisions(s.meta)
}

func (s *OssStorage) GetStackBoundRevisions(stack string) []uint64 {
	return getStackBoundRevisions(s.meta, stack)
}

func (s *OssStorage) GetLatestRevision() uint64 {
	return s.meta.LatestRevision
}

func (s *OssStorage) Create(r *v1.Release) error {
	if checkRevisionExistence(s.meta, r.Revision) {
		return ErrReleaseAlreadyExist
	}

	if err := s.writeRelease(r); err != nil {
		return err
	}

	addLatestReleaseMetaData(s.meta, r.Revision, r.Stack, r.Phase)
	return s.writeMeta()
}

func (s *OssStorage) Update(r *v1.Release) error {
	if !checkRevisionExistence(s.meta, r.Revision) {
		return ErrReleaseNotExist
	}

	return s.writeRelease(r)
}

func (s *OssStorage) readMeta() error {
	body, err := s.bucket.GetObject(s.prefix + "/" + metadataFile)
	if err != nil {
		ossErr, ok := err.(oss.ServiceError)
		// error code ref: github.com/aliyun/aliyun-oss-go-sdk@v2.1.8+incompatible/oss/bucket.go:553
		if ok && ossErr.StatusCode == 404 {
			s.meta = &releasesMetaData{}
			return nil
		}
		return fmt.Errorf("get releases metadata from oss failed: %w", err)
	}
	defer func() {
		_ = body.Close()
	}()

	content, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("read releases metadata failed: %w", err)
	}
	if len(content) == 0 {
		s.meta = &releasesMetaData{}
		return nil
	}

	meta := &releasesMetaData{}
	if err = yaml.Unmarshal(content, meta); err != nil {
		return fmt.Errorf("yaml unmarshal releases metadata failed: %w", err)
	}
	s.meta = meta
	return nil
}

func (s *OssStorage) writeMeta() error {
	content, err := yaml.Marshal(s.meta)
	if err != nil {
		return fmt.Errorf("yaml marshal releases metadata failed: %w", err)
	}

	if err = s.bucket.PutObject(s.prefix+"/"+metadataFile, bytes.NewReader(content)); err != nil {
		return fmt.Errorf("put releases metadata to oss failed: %w", err)
	}
	return nil
}

func (s *OssStorage) writeRelease(r *v1.Release) error {
	content, err := yaml.Marshal(r)
	if err != nil {
		return fmt.Errorf("yaml marshal release failed: %w", err)
	}

	key := fmt.Sprintf("%s/%d%s", s.prefix, r.Revision, yamlSuffix)
	if err = s.bucket.PutObject(key, bytes.NewReader(content)); err != nil {
		return fmt.Errorf("put release to oss failed: %w", err)
	}
	return nil
}

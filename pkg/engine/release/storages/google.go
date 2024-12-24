package storages

import (
	"context"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"

	googlestorage "cloud.google.com/go/storage"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

// GoogleStorage is an implementation of release.Storage which uses google cloud as storage.
type GoogleStorage struct {
	bucket googlestorage.BucketHandle

	// The prefix to store the release files.
	prefix string

	meta *releasesMetaData
}

// NewGoogleStorage news google cloud release storage, and derives metadata.
func NewGoogleStorage(bucket *googlestorage.BucketHandle, prefix string) (*GoogleStorage, error) {
	s := &GoogleStorage{
		bucket: *bucket,
		prefix: prefix,
	}
	if err := s.readMeta(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *GoogleStorage) Get(revision uint64) (*v1.Release, error) {
	ctx := context.Background()
	if !checkRevisionExistence(s.meta, revision) {
		return nil, ErrReleaseNotExist
	}

	obj := s.bucket.Object(fmt.Sprintf("%s/%d%s", s.prefix, revision, yamlSuffix))
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("get release from google storage failed: %w", err)
	}
	defer reader.Close()
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read release failed: %w", err)
	}

	rel := &v1.Release{}
	if err = yaml.Unmarshal(content, rel); err != nil {
		return nil, fmt.Errorf("yaml unmarshal release failed: %w", err)
	}
	return rel, nil
}

func (s *GoogleStorage) GetRevisions() []uint64 {
	return getRevisions(s.meta)
}

func (s *GoogleStorage) GetStackBoundRevisions(stack string) []uint64 {
	return getStackBoundRevisions(s.meta, stack)
}

func (s *GoogleStorage) GetLatestRevision() uint64 {
	return s.meta.LatestRevision
}

func (s *GoogleStorage) Create(r *v1.Release) error {
	if checkRevisionExistence(s.meta, r.Revision) {
		return ErrReleaseAlreadyExist
	}

	if err := s.writeRelease(r); err != nil {
		return err
	}

	addLatestReleaseMetaData(s.meta, r.Revision, r.Stack)
	return s.writeMeta()
}

func (s *GoogleStorage) Update(r *v1.Release) error {
	if !checkRevisionExistence(s.meta, r.Revision) {
		return ErrReleaseNotExist
	}

	return s.writeRelease(r)
}

func (s *GoogleStorage) readMeta() error {
	ctx := context.Background()
	obj := s.bucket.Object(s.prefix + "/" + metadataFile)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		if err == googlestorage.ErrObjectNotExist {
			s.meta = &releasesMetaData{}
			return nil
		}
		return fmt.Errorf("get releases metadata from google failed: %w", err)
	}
	defer reader.Close()
	content, err := io.ReadAll(reader)
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

func (s *GoogleStorage) writeMeta() error {
	ctx := context.Background()
	obj := s.bucket.Object(s.prefix + "/" + metadataFile)
	content, err := yaml.Marshal(s.meta)
	if err != nil {
		return fmt.Errorf("yaml marshal releases metadata failed: %w", err)
	}

	writer := obj.NewWriter(ctx)
	if _, err = writer.Write(content); err != nil {
		return fmt.Errorf("write releases metadata failed: %w", err)
	}

	if err = writer.Close(); err != nil {
		return fmt.Errorf("close writer failed: %w", err)
	}
	return nil
}

func (s *GoogleStorage) writeRelease(r *v1.Release) error {
	content, err := yaml.Marshal(r)
	if err != nil {
		return fmt.Errorf("yaml marshal release failed: %w", err)
	}

	obj := s.bucket.Object(fmt.Sprintf("%s/%d%s", s.prefix, r.Revision, yamlSuffix))
	writer := obj.NewWriter(context.Background())
	if _, err = writer.Write(content); err != nil {
		return fmt.Errorf("write release failed: %w", err)
	}

	if err = writer.Close(); err != nil {
		return fmt.Errorf("close writer failed: %w", err)
	}
	return nil
}

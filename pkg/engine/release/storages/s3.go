package storages

import (
	"bytes"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

// S3Storage is an implementation of release.Storage which uses s3 as storage.
type S3Storage struct {
	s3     *s3.S3
	bucket string

	// The prefix to store the release files.
	prefix string

	meta *releasesMetaData
}

// NewS3Storage news s3 release storage, and derives metadata.
func NewS3Storage(s3 *s3.S3, bucket, prefix string) (*S3Storage, error) {
	s := &S3Storage{
		s3:     s3,
		bucket: bucket,
		prefix: prefix,
	}
	if err := s.readMeta(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *S3Storage) Get(revision uint64) (*v1.Release, error) {
	if !checkRevisionExistence(s.meta, revision) {
		return nil, ErrReleaseNotExist
	}

	key := fmt.Sprintf("%s/%d%s", s.prefix, revision, yamlSuffix)
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    &key,
	}
	output, err := s.s3.GetObject(input)
	if err != nil {
		return nil, fmt.Errorf("get release from s3 failed: %w", err)
	}
	defer func() {
		_ = output.Body.Close()
	}()
	content, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("read release failed: %w", err)
	}

	r := &v1.Release{}
	if err = yaml.Unmarshal(content, r); err != nil {
		return nil, fmt.Errorf("yaml unmarshal release failed: %w", err)
	}
	return r, nil
}

func (s *S3Storage) GetRevisions() []uint64 {
	return getRevisions(s.meta)
}

func (s *S3Storage) GetStackBoundRevisions(stack string) []uint64 {
	return getStackBoundRevisions(s.meta, stack)
}

func (s *S3Storage) GetLatestRevision() uint64 {
	return s.meta.LatestRevision
}

func (s *S3Storage) Create(r *v1.Release) error {
	if checkRevisionExistence(s.meta, r.Revision) {
		return ErrReleaseAlreadyExist
	}

	if err := s.writeRelease(r); err != nil {
		return err
	}

	addLatestReleaseMetaData(s.meta, r.Revision, r.Stack, r.Phase)
	return s.writeMeta()
}

func (s *S3Storage) Update(r *v1.Release) error {
	if !checkRevisionExistence(s.meta, r.Revision) {
		return ErrReleaseNotExist
	}

	return s.writeRelease(r)
}

func (s *S3Storage) readMeta() error {
	key := s.prefix + "/" + metadataFile
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    &key,
	}
	output, err := s.s3.GetObject(input)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == s3.ErrCodeNoSuchKey {
			s.meta = &releasesMetaData{}
			return nil
		}
		return fmt.Errorf("get releases metadata from s3 failed: %w", err)
	}
	defer func() {
		_ = output.Body.Close()
	}()

	content, err := io.ReadAll(output.Body)
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

func (s *S3Storage) writeMeta() error {
	content, err := yaml.Marshal(s.meta)
	if err != nil {
		return fmt.Errorf("yaml marshal releases metadata failed: %w", err)
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.prefix + "/" + metadataFile),
		Body:   bytes.NewReader(content),
	}
	if _, err = s.s3.PutObject(input); err != nil {
		return fmt.Errorf("put releases metadata to s3 failed: %w", err)
	}
	return nil
}

func (s *S3Storage) writeRelease(r *v1.Release) error {
	content, err := yaml.Marshal(r)
	if err != nil {
		return fmt.Errorf("yaml marshal release failed: %w", err)
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fmt.Sprintf("%s/%d%s", s.prefix, r.Revision, yamlSuffix)),
		Body:   bytes.NewReader(content),
	}
	if _, err = s.s3.PutObject(input); err != nil {
		return fmt.Errorf("put release to s3 failed: %w", err)
	}
	return nil
}

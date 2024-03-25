package storages

import (
	"bytes"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/engine/spec"
)

// S3Storage should implement the spec.Storage interface.
var _ spec.Storage = &S3Storage{}

type S3Storage struct {
	s3     *s3.S3
	bucket string

	// The s3 key to store the state file.
	key string
}

// NewS3Storage constructs an AWS S3 based spec storage.
func NewS3Storage(s3 *s3.S3, bucket, key string) *S3Storage {
	return &S3Storage{
		s3:     s3,
		bucket: bucket,
		key:    key,
	}
}

// Get returns the Spec, if the Spec does not exist, return nil.
func (s *S3Storage) Get() (*v1.Intent, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    &s.key,
	}
	output, err := s.s3.GetObject(input)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		// if no kusion intent file, return nil intent
		if ok && awsErr.Code() == s3.ErrCodeNoSuchKey {
			return nil, nil
		}
		return nil, err
	}
	defer func() {
		_ = output.Body.Close()
	}()

	content, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, err
	}
	if len(content) == 0 {
		return nil, nil
	}

	intent := &v1.Intent{}
	err = yaml.Unmarshal(content, intent)
	if err != nil {
		return nil, err
	}
	return intent, nil
}

// Apply updates the spec if already exists, or create a new spec.
func (s *S3Storage) Apply(intent *v1.Intent) error {
	content, err := yaml.Marshal(intent)
	if err != nil {
		return err
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key),
		Body:   bytes.NewReader(content),
	}
	_, err = s.s3.PutObject(input)
	return err
}

package storages

import (
	"bytes"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

// S3Storage is an implementation of state.Storage which uses s3 as storage.
type S3Storage struct {
	s3     *s3.S3
	bucket string

	// The s3 key to store the state file.
	key string
}

func NewS3Storage(s3 *s3.S3, bucket, key string) *S3Storage {
	return &S3Storage{
		s3:     s3,
		bucket: bucket,
		key:    key,
	}
}

func (s *S3Storage) Get() (*v1.State, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    &s.key,
	}
	output, err := s.s3.GetObject(input)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		// if no kusion state file, return nil state
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

	state := &v1.State{}
	err = yaml.Unmarshal(content, state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (s *S3Storage) Apply(state *v1.State) error {
	content, err := yaml.Marshal(state)
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

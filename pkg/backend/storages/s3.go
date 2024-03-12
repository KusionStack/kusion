package storages

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/engine/state"
	statestorages "kusionstack.io/kusion/pkg/engine/state/storages"
)

// S3Storage is an implementation of backend.Backend which uses s3 as storage.
type S3Storage struct {
	s3     *s3.S3
	bucket string

	// prefix will be added to the object storage key, so that all the files are stored under the prefix.
	prefix string
}

func NewS3Storage(config *v1.BackendS3Config) (*S3Storage, error) {
	c := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(config.AccessKeyID, config.AccessKeySecret, ""),
		Region:           aws.String(config.Region),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(false),
	}
	if config.Endpoint != "" {
		c.Endpoint = aws.String(config.Endpoint)
	}
	sess, err := session.NewSession(c)
	if err != nil {
		return nil, err
	}

	return &S3Storage{
		s3:     s3.New(sess),
		bucket: config.Bucket,
		prefix: config.Prefix,
	}, nil
}

func (s *S3Storage) StateStorage(project, stack, workspace string) state.Storage {
	return statestorages.NewS3Storage(s.s3, s.bucket, statestorages.GenGenericOssStateFileKey(s.prefix, project, stack, workspace))
}

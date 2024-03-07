package storages

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

// S3Storage is an implementation of backend.Backend which uses s3 as storage.
type S3Storage struct {
	sess   *session.Session
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
		sess:   sess,
		bucket: config.Bucket,
		prefix: config.Prefix,
	}, nil
}

package s3

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/zclconf/go-cty/cty"
	"kusionstack.io/kusion/pkg/engine/states"
)

type S3Backend struct {
	S3State
}

func NewS3Backend() states.Backend {
	return &S3Backend{}
}

// ConfigSchema returns a description of the expected configuration
// structure for the receiving backend.
func (b *S3Backend) ConfigSchema() cty.Type {
	config := map[string]cty.Type{
		"endpoint":        cty.String,
		"bucket":          cty.String,
		"accessKeyID":     cty.String,
		"accessKeySecret": cty.String,
		"region":          cty.String,
	}
	return cty.Object(config)
}

// Configure uses the provided configuration to set configuration fields
// within the S3State backend.
func (b *S3Backend) Configure(obj cty.Value) error {
	var endpoint, bucket, accessKeyID, accessKeySecret, region cty.Value
	if endpoint = obj.GetAttr("endpoint"); endpoint.IsNull() {
		return errors.New("s3 endpoint must be configure in backend config")
	}
	if bucket = obj.GetAttr("bucket"); bucket.IsNull() {
		return errors.New("s3 bucket must be configure in backend config")
	}
	if accessKeyID = obj.GetAttr("accessKeyID"); bucket.IsNull() {
		return errors.New("s3 accessKeyID must be configure in backend config")
	}
	if accessKeySecret = obj.GetAttr("accessKeySecret"); accessKeySecret.IsNull() {
		return errors.New("s3 accessKeySecret must be configure in backend config")
	}
	if region = obj.GetAttr("region"); region.IsNull() {
		return errors.New("s3 region must be configure in backend config")
	}
	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKeyID.AsString(), accessKeySecret.AsString(), ""),
		Endpoint:         aws.String(endpoint.AsString()),
		Region:           aws.String(region.AsString()),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return err
	}
	s3State := &S3State{
		sess:       sess,
		bucketName: bucket.AsString(),
	}
	b.S3State = *s3State
	return nil
}

// StateStorage return a StateStorage to manage State stored in S3
func (b *S3Backend) StateStorage() states.StateStorage {
	return &S3State{b.sess, b.bucketName}
}

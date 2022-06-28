//go:build !arm64
// +build !arm64

package remote

import (
	"encoding/json"
	"testing"
	"time"

	"kusionstack.io/kusion/pkg/engine/states"

	"bou.ke/monkey"

	"github.com/Azure/go-autorest/autorest/mocks"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
)

func S3StateSetUp(t *testing.T) *S3State {
	sess := &session.Session{}
	bucketName := "test_bucket"

	monkey.Patch(s3.New, func(p client.ConfigProvider, cfgs ...*aws.Config) *s3.S3 {
		return &s3.S3{}
	})
	monkey.Patch((*s3.S3).PutObject, func(c *s3.S3, input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
		return nil, nil
	})

	monkey.Patch((*s3.S3).ListObjects, func(c *s3.S3, input *s3.ListObjectsInput) (*s3.ListObjectsOutput, error) {
		return &s3.ListObjectsOutput{Contents: []*s3.Object{{LastModified: aws.Time(time.Now())}}}, nil
	})
	state := &states.State{Tenant: "test_global_tenant", Project: "test_project", Stack: "test_env"}
	jsonByte, _ := json.MarshalIndent(state, "", "  ")
	monkey.Patch((*s3.S3).GetObject, func(c *s3.S3, input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
		return &s3.GetObjectOutput{Body: mocks.NewBody(string(jsonByte))}, nil
	})
	monkey.Patch(session.NewSession, func(cfgs ...*aws.Config) (*session.Session, error) {
		return &session.Session{}, nil
	})

	return &S3State{sess: sess, bucketName: bucketName}
}

func TestS3State(t *testing.T) {
	defer monkey.UnpatchAll()
	s3State := S3StateSetUp(t)

	_, err := NewS3State("test_endpoint", "test_access_key", "test_access_secret", "test_bucket", "test_region")
	assert.NoError(t, err)
	state := &states.State{Tenant: "test_global_tenant", Project: "test_project", Stack: "test_env"}
	err = s3State.Apply(state)
	assert.NoError(t, err)
	query := &states.StateQuery{Tenant: "test_global_tenant", Project: "test_project", Stack: "test_env"}
	latestState, err := s3State.GetLatestState(query)
	assert.NoError(t, err)
	assert.Equal(t, state, latestState)

	defer func() {
		if r := recover(); r != "implement me" {
			t.Errorf("Delete() got: %v, want: 'implement me'", r)
		}
	}()
	s3State.Delete("test")
}

package s3

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Azure/go-autorest/autorest/mocks"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/engine/states"
)

func S3StateSetUp(t *testing.T) *S3State {
	sess := &session.Session{}
	bucketName := "test_bucket"

	mockey.Mock(s3.New).To(func(p client.ConfigProvider, cfgs ...*aws.Config) *s3.S3 {
		return &s3.S3{}
	}).Build()
	mockey.Mock((*s3.S3).PutObject).To(func(c *s3.S3, input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
		return nil, nil
	}).Build()

	mockey.Mock((*s3.S3).ListObjects).To(func(c *s3.S3, input *s3.ListObjectsInput) (*s3.ListObjectsOutput, error) {
		return &s3.ListObjectsOutput{Contents: []*s3.Object{{LastModified: aws.Time(time.Now())}}}, nil
	}).Build()
	state := &states.State{Tenant: "test_global_tenant", Project: "test_project", Stack: "test_env"}
	jsonByte, _ := json.MarshalIndent(state, "", "  ")
	mockey.Mock((*s3.S3).GetObject).To(func(c *s3.S3, input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
		return &s3.GetObjectOutput{Body: mocks.NewBody(string(jsonByte))}, nil
	}).Build()
	mockey.Mock(session.NewSession).To(func(cfgs ...*aws.Config) (*session.Session, error) {
		return &session.Session{}, nil
	}).Build()

	return &S3State{sess: sess, bucketName: bucketName}
}

func TestS3State(t *testing.T) {
	mockey.PatchConvey("test s3 state", t, func() {
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
	})
}

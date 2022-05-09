package states

import (
	"encoding/json"
	"io"
	"testing"
	"time"

	"bou.ke/monkey"

	"github.com/Azure/go-autorest/autorest/mocks"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/stretchr/testify/assert"
)

func SetUp(t *testing.T) *OssState {
	bucket := &oss.Bucket{}

	monkey.Patch(oss.New, func(endpoint, accessKeyID, accessKeySecret string, options ...oss.ClientOption) (*oss.Client, error) {
		return &oss.Client{}, nil

	})

	monkey.Patch(oss.Bucket.PutObject, func(b oss.Bucket, objectKey string, reader io.Reader, options ...oss.Option) error {
		return nil
	})
	monkey.Patch(oss.Bucket.ListObjects, func(b oss.Bucket, options ...oss.Option) (oss.ListObjectsResult, error) {
		return oss.ListObjectsResult{Objects: []oss.ObjectProperties{{LastModified: time.Now()}}}, nil
	})
	state := &State{Tenant: "test_global_tenant", Project: "test_project", Stack: "test_env"}
	jsonByte, _ := json.MarshalIndent(state, "", "  ")
	monkey.Patch(oss.Bucket.GetObject, func(b oss.Bucket, objectKey string, options ...oss.Option) (io.ReadCloser, error) {
		return mocks.NewBody(string(jsonByte)), nil
	})

	return &OssState{bucket: bucket}
}

func TestOssState(t *testing.T) {
	defer monkey.UnpatchAll()
	ossState := SetUp(t)
	_, err := NewOSSState("test_endpoint", "test_access_id", "test_access_secret", "testbucket")
	assert.NoError(t, err)
	state := &State{Tenant: "test_global_tenant", Project: "test_project", Stack: "test_env"}
	err = ossState.Apply(state)
	assert.NoError(t, err)
	query := &StateQuery{Tenant: "test_global_tenant", Project: "test_project", Stack: "test_env"}
	latestState, err := ossState.GetLatestState(query)
	assert.NoError(t, err)
	assert.Equal(t, state, latestState)

	defer func() {
		if r := recover(); r != "implement me" {
			t.Errorf("Delete() got: %v, want: 'implement me'", r)
		}
	}()
	ossState.Delete("test")
}

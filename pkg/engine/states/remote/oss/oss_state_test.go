//go:build !arm64
// +build !arm64

package oss

import (
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/Azure/go-autorest/autorest/mocks"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/engine/states"
)

func SetUp(t *testing.T) *OssState {
	bucket := &oss.Bucket{}

	mockey.Mock(oss.New).To(func(endpoint, accessKeyID, accessKeySecret string, options ...oss.ClientOption) (*oss.Client, error) {
		return &oss.Client{}, nil
	}).Build()

	mockey.Mock(oss.Bucket.PutObject).To(func(b oss.Bucket, objectKey string, reader io.Reader, options ...oss.Option) error {
		return nil
	}).Build()
	mockey.Mock(oss.Bucket.ListObjects).To(func(b oss.Bucket, options ...oss.Option) (oss.ListObjectsResult, error) {
		return oss.ListObjectsResult{Objects: []oss.ObjectProperties{{LastModified: time.Now()}}}, nil
	}).Build()
	state := &states.State{Tenant: "test_global_tenant", Project: "test_project", Stack: "test_env"}
	jsonByte, _ := json.MarshalIndent(state, "", "  ")
	mockey.Mock(oss.Bucket.GetObject).To(func(b oss.Bucket, objectKey string, options ...oss.Option) (io.ReadCloser, error) {
		return mocks.NewBody(string(jsonByte)), nil
	}).Build()

	return &OssState{bucket: bucket}
}

func TestOssState(t *testing.T) {
	mockey.PatchConvey("test oss state", t, func() {
		ossState := SetUp(t)
		_, err := NewOSSState("test_endpoint", "test_access_id", "test_access_secret", "testbucket")
		assert.NoError(t, err)
		state := &states.State{Tenant: "test_global_tenant", Project: "test_project", Stack: "test_env"}
		err = ossState.Apply(state)
		assert.NoError(t, err)
		query := &states.StateQuery{Tenant: "test_global_tenant", Project: "test_project", Stack: "test_env"}
		latestState, err := ossState.GetLatestState(query)
		assert.NoError(t, err)
		assert.Equal(t, state, latestState)

		defer func() {
			if r := recover(); r != "implement me" {
				t.Errorf("Delete() got: %v, want: 'implement me'", r)
			}
		}()
		ossState.Delete("test")
	})
}

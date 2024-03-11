package storages

import (
	"bytes"
	"io"
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

func mockOssStorage() *OssStorage {
	return &OssStorage{bucket: &oss.Bucket{}}
}

func TestOssStorage_Get(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		content []byte
		state   *v1.State
	}{
		{
			name:    "get oss state successfully",
			success: true,
			content: []byte(mockStateContent()),
			state:   mockState(),
		},
		{
			name:    "get empty oss state successfully",
			success: true,
			content: nil,
			state:   nil,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock oss get", t, func() {
				mockey.Mock(oss.Bucket.GetObject).Return(io.NopCloser(bytes.NewReader([]byte(""))), nil).Build()
				mockey.Mock(io.ReadAll).Return(tc.content, nil).Build()
				state, err := mockOssStorage().Get()
				assert.Equal(t, tc.success, err == nil)
				assert.Equal(t, tc.state, state)
			})
		})
	}
}

func TestOssStorage_Apply(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		state   *v1.State
	}{
		{
			name:    "apply oss state successfully",
			success: true,
			state:   mockState(),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock oss put", t, func() {
				mockey.Mock(oss.Bucket.PutObject).Return(nil).Build()
				err := mockOssStorage().Apply(tc.state)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

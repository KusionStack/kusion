package storages

import (
	"bytes"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

func mockS3Storage() *S3Storage {
	return &S3Storage{s3: &s3.S3{}}
}

func TestS3Storage_Get(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		content []byte
		state   *v1.DeprecatedState
	}{
		{
			name:    "get s3 state successfully",
			success: true,
			content: []byte(mockStateContent()),
			state:   mockState(),
		},
		{
			name:    "get empty s3 state successfully",
			success: true,
			content: nil,
			state:   nil,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock s3 get", t, func() {
				mockey.Mock((*s3.S3).GetObject).Return(&s3.GetObjectOutput{
					Body: io.NopCloser(bytes.NewReader([]byte(""))),
				}, nil).Build()
				mockey.Mock(io.ReadAll).Return(tc.content, nil).Build()
				state, err := mockS3Storage().Get()
				assert.Equal(t, tc.success, err == nil)
				assert.Equal(t, tc.state, state)
			})
		})
	}
}

func TestS3Storage_Apply(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		state   *v1.DeprecatedState
	}{
		{
			name:    "apply s3 state successfully",
			success: true,
			state:   mockState(),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock s3 put", t, func() {
				mockey.Mock((*s3.S3).PutObject).Return(nil, nil).Build()
				err := mockS3Storage().Apply(tc.state)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

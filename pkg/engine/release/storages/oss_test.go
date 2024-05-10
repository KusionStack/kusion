package storages

import (
	"bytes"
	"io"
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

func mockOssStorage() *OssStorage {
	return &OssStorage{bucket: &oss.Bucket{}, meta: mockReleasesMeta()}
}

func mockOssStorageWriteMeta() {
	mockey.Mock((*OssStorage).writeMeta).Return(nil).Build()
}

func mockOssStorageWriteRelease() {
	mockey.Mock((*OssStorage).writeRelease).Return(nil).Build()
}

func TestOssStorage_Get(t *testing.T) {
	testcases := []struct {
		name            string
		success         bool
		revision        uint64
		content         []byte
		expectedRelease *v1.Release
	}{
		{
			name:            "get release successfully",
			success:         true,
			revision:        1,
			content:         []byte(mockReleaseRevision1Content()),
			expectedRelease: mockRelease(1),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock oss operation", t, func() {
				mockey.Mock(oss.Bucket.GetObject).Return(io.NopCloser(bytes.NewReader([]byte(""))), nil).Build()
				mockey.Mock(io.ReadAll).Return(tc.content, nil).Build()
				r, err := mockOssStorage().Get(tc.revision)
				assert.Equal(t, tc.success, err == nil)
				if tc.success {
					expectedReleaseContent, _ := yaml.Marshal(tc.expectedRelease)
					releaseContent, _ := yaml.Marshal(r)
					assert.Equal(t, string(expectedReleaseContent), string(releaseContent))
				}
			})
		})
	}
}

func TestOssStorage_GetRevisions(t *testing.T) {
	testcases := []struct {
		name              string
		expectedRevisions []uint64
	}{
		{
			name:              "get release revisions successfully",
			expectedRevisions: []uint64{1, 2, 3},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock oss operation", t, func() {
				revisions := mockOssStorage().GetRevisions()
				assert.Equal(t, tc.expectedRevisions, revisions)
			})
		})
	}
}

func TestOssStorage_GetStackBoundRevisions(t *testing.T) {
	testcases := []struct {
		name              string
		stack             string
		expectedRevisions []uint64
	}{
		{
			name:              "get stack bound release revisions successfully",
			stack:             "test_stack",
			expectedRevisions: []uint64{1, 2, 3},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock oss operation", t, func() {
				revisions := mockOssStorage().GetStackBoundRevisions(tc.stack)
				assert.Equal(t, tc.expectedRevisions, revisions)
			})
		})
	}
}

func TestOssStorage_GetLatestRevision(t *testing.T) {
	testcases := []struct {
		name             string
		expectedRevision uint64
	}{
		{
			name:             "get latest release revision successfully",
			expectedRevision: 3,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock oss operation", t, func() {
				revision := mockOssStorage().GetLatestRevision()
				assert.Equal(t, tc.expectedRevision, revision)
			})
		})
	}
}

func TestOssStorage_Create(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		r       *v1.Release
	}{
		{
			name:    "create release successfully",
			success: true,
			r:       mockRelease(4),
		},
		{
			name:    "failed to create release already exist",
			success: false,
			r:       mockRelease(3),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock oss operation", t, func() {
				mockOssStorageWriteMeta()
				mockOssStorageWriteRelease()
				err := mockOssStorage().Create(tc.r)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestOssStorage_Update(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		r       *v1.Release
	}{
		{
			name:    "update release successfully",
			success: true,
			r:       mockRelease(3),
		},
		{
			name:    "failed to update release not exist",
			success: false,
			r:       mockRelease(4),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock oss operation", t, func() {
				mockOssStorageWriteRelease()
				err := mockOssStorage().Update(tc.r)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

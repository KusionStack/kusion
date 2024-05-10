package storages

import (
	"bytes"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

func mockS3Storage() *S3Storage {
	return &S3Storage{s3: &s3.S3{}, meta: mockReleasesMeta()}
}

func mockS3StorageWriteMeta() {
	mockey.Mock((*S3Storage).writeMeta).Return(nil).Build()
}

func mockS3StorageWriteRelease() {
	mockey.Mock((*S3Storage).writeRelease).Return(nil).Build()
}

func TestS3Storage_Get(t *testing.T) {
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
			mockey.PatchConvey("mock s3 operation", t, func() {
				mockey.Mock((*s3.S3).GetObject).Return(&s3.GetObjectOutput{
					Body: io.NopCloser(bytes.NewReader([]byte(""))),
				}, nil).Build()
				mockey.Mock(io.ReadAll).Return(tc.content, nil).Build()
				r, err := mockS3Storage().Get(tc.revision)
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

func TestS3Storage_GetRevisions(t *testing.T) {
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
			mockey.PatchConvey("mock s3 operation", t, func() {
				revisions := mockS3Storage().GetRevisions()
				assert.Equal(t, tc.expectedRevisions, revisions)
			})
		})
	}
}

func TestS3Storage_GetStackBoundRevisions(t *testing.T) {
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
			mockey.PatchConvey("mock s3 operation", t, func() {
				revisions := mockS3Storage().GetStackBoundRevisions(tc.stack)
				assert.Equal(t, tc.expectedRevisions, revisions)
			})
		})
	}
}

func TestS3Storage_GetLatestRevision(t *testing.T) {
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
			mockey.PatchConvey("mock s3 operation", t, func() {
				revision := mockS3Storage().GetLatestRevision()
				assert.Equal(t, tc.expectedRevision, revision)
			})
		})
	}
}

func TestS3Storage_Create(t *testing.T) {
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
			mockey.PatchConvey("mock s3 operation", t, func() {
				mockS3StorageWriteMeta()
				mockS3StorageWriteRelease()
				err := mockS3Storage().Create(tc.r)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestS3Storage_Update(t *testing.T) {
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
			mockey.PatchConvey("mock s3 operation", t, func() {
				mockS3StorageWriteRelease()
				err := mockS3Storage().Update(tc.r)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

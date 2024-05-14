package release

import (
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/engine/release/storages"
)

func mockReleaseStorageOperation(revision uint64) {
	mockey.Mock((*storages.LocalStorage).GetLatestRevision).Return(revision).Build()
	mockey.Mock((*storages.LocalStorage).Get).Return(&v1.Release{State: &v1.State{}}, nil).Build()
}

func TestGetLatestState(t *testing.T) {
	testcases := []struct {
		name             string
		success          bool
		revision         uint64
		expectedNilState bool
	}{
		{
			name:             "nil release",
			success:          true,
			revision:         0,
			expectedNilState: true,
		},
		{
			name:             "not nil release",
			success:          true,
			revision:         1,
			expectedNilState: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock s3 operation", t, func() {
				mockReleaseStorageOperation(tc.revision)
				state, err := GetLatestState(&storages.LocalStorage{})
				assert.Equal(t, tc.success, err == nil)
				assert.Equal(t, tc.expectedNilState, state == nil)
			})
		})
	}
}

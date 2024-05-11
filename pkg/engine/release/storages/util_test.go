package storages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

func mockReleasesMetaData() *releasesMetaData {
	return &releasesMetaData{
		LatestRevision: 3,
		ReleaseMetaDatas: []*releaseMetaData{
			{
				Revision: 1,
				Stack:    "dev",
				Phase:    v1.ReleasePhaseSucceeded,
			},
			{
				Revision: 2,
				Stack:    "pre",
				Phase:    v1.ReleasePhaseFailed,
			},
			{
				Revision: 3,
				Stack:    "pre",
				Phase:    v1.ReleasePhaseSucceeded,
			},
		},
	}
}

func TestCheckReleaseExistence(t *testing.T) {
	testcases := []struct {
		name     string
		meta     *releasesMetaData
		revision uint64
		exist    bool
	}{
		{
			name:     "empty releases meta data",
			meta:     &releasesMetaData{},
			revision: 2,
			exist:    false,
		},
		{
			name:     "exist workspace",
			meta:     mockReleasesMetaData(),
			revision: 2,
			exist:    true,
		},
		{
			name:     "not exist workspace",
			meta:     mockReleasesMetaData(),
			revision: 5,
			exist:    false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			exist := checkRevisionExistence(tc.meta, tc.revision)
			assert.Equal(t, tc.exist, exist)
		})
	}
}

func TestGetRevisions(t *testing.T) {
	testcases := []struct {
		name              string
		meta              *releasesMetaData
		expectedRevisions []uint64
	}{
		{
			name:              "get revisions",
			meta:              mockReleasesMetaData(),
			expectedRevisions: []uint64{1, 2, 3},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			revisions := getRevisions(tc.meta)
			assert.Equal(t, tc.expectedRevisions, revisions)
		})
	}
}

func TestGetStackBoundRevisions(t *testing.T) {
	testcases := []struct {
		name              string
		meta              *releasesMetaData
		stack             string
		expectedRevisions []uint64
	}{
		{
			name:              "get stack bound revisions",
			meta:              mockReleasesMetaData(),
			stack:             "pre",
			expectedRevisions: []uint64{2, 3},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			revisions := getStackBoundRevisions(tc.meta, tc.stack)
			assert.Equal(t, tc.expectedRevisions, revisions)
		})
	}
}

func TestAddLatestReleaseMetaData(t *testing.T) {
	testcases := []struct {
		name         string
		meta         *releasesMetaData
		revision     uint64
		stack        string
		phase        v1.ReleasePhase
		expectedMeta *releasesMetaData
	}{
		{
			name:     "empty releases meta data add release",
			meta:     &releasesMetaData{},
			revision: 1,
			stack:    "prod",
			phase:    v1.ReleasePhaseGenerating,
			expectedMeta: &releasesMetaData{
				LatestRevision: 1,
				ReleaseMetaDatas: []*releaseMetaData{
					{
						Revision: 1,
						Stack:    "prod",
						Phase:    v1.ReleasePhaseGenerating,
					},
				},
			},
		},
		{
			name:     "non-empty releases meta data add release",
			meta:     mockReleasesMetaData(),
			revision: 4,
			stack:    "prod",
			phase:    v1.ReleasePhasePreviewing,
			expectedMeta: &releasesMetaData{
				LatestRevision: 4,
				ReleaseMetaDatas: []*releaseMetaData{
					{
						Revision: 1,
						Stack:    "dev",
						Phase:    v1.ReleasePhaseSucceeded,
					},
					{
						Revision: 2,
						Stack:    "pre",
						Phase:    v1.ReleasePhaseFailed,
					},
					{
						Revision: 3,
						Stack:    "pre",
						Phase:    v1.ReleasePhaseSucceeded,
					},
					{
						Revision: 4,
						Stack:    "prod",
						Phase:    v1.ReleasePhasePreviewing,
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			addLatestReleaseMetaData(tc.meta, tc.revision, tc.stack, tc.phase)
			assert.Equal(t, tc.expectedMeta, tc.expectedMeta)
		})
	}
}

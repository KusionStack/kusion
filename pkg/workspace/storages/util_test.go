package storages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockWorkspacesMetaData() *workspacesMetaData {
	return &workspacesMetaData{
		Current: "dev",
		AvailableWorkspaces: []string{
			"default",
			"dev",
			"prod",
		},
	}
}

func TestCheckWorkspaceExistence(t *testing.T) {
	testcases := []struct {
		name   string
		meta   *workspacesMetaData
		wsName string
		exist  bool
	}{
		{
			name:   "empty workspaces meta data",
			meta:   &workspacesMetaData{},
			wsName: "dev",
			exist:  false,
		},
		{
			name:   "exist workspace",
			meta:   mockWorkspacesMetaData(),
			wsName: "dev",
			exist:  true,
		},
		{
			name:   "not exist workspace",
			meta:   mockWorkspacesMetaData(),
			wsName: "pre",
			exist:  false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			exist := checkWorkspaceExistence(tc.meta, tc.wsName)
			assert.Equal(t, tc.exist, exist)
		})
	}
}

func TestAddAvailableWorkspaces(t *testing.T) {
	testcases := []struct {
		name         string
		meta         *workspacesMetaData
		wsName       string
		expectedMeta *workspacesMetaData
	}{
		{
			name:   "empty workspaces meta data add workspace",
			meta:   &workspacesMetaData{},
			wsName: "default",
			expectedMeta: &workspacesMetaData{
				AvailableWorkspaces: []string{"default"},
			},
		},
		{
			name:   "non-empty workspaces meta data add workspace",
			meta:   mockWorkspacesMetaData(),
			wsName: "pre",
			expectedMeta: &workspacesMetaData{
				Current: "dev",
				AvailableWorkspaces: []string{
					"default",
					"dev",
					"prod",
					"pre",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			addAvailableWorkspaces(tc.meta, tc.wsName)
			assert.Equal(t, tc.expectedMeta, tc.expectedMeta)
		})
	}
}

func TestRemoveAvailableWorkspaces(t *testing.T) {
	testcases := []struct {
		name         string
		meta         *workspacesMetaData
		wsName       string
		expectedMeta *workspacesMetaData
	}{
		{
			name:   "remove not exist workspace",
			meta:   mockWorkspacesMetaData(),
			wsName: "pre",
			expectedMeta: &workspacesMetaData{
				Current: "dev",
				AvailableWorkspaces: []string{
					"default",
					"dev",
					"prod",
				},
			},
		},
		{
			name:   "remove exist workspace",
			meta:   mockWorkspacesMetaData(),
			wsName: "prod",
			expectedMeta: &workspacesMetaData{
				Current: "dev",
				AvailableWorkspaces: []string{
					"default",
					"dev",
				},
			},
		},
		{
			name:   "set current workspace",
			meta:   mockWorkspacesMetaData(),
			wsName: "dev",
			expectedMeta: &workspacesMetaData{
				Current: "default",
				AvailableWorkspaces: []string{
					"default",
					"prod",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			removeAvailableWorkspaces(tc.meta, tc.wsName)
			assert.Equal(t, tc.expectedMeta, tc.expectedMeta)
		})
	}
}

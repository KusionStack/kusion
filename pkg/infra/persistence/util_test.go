package persistence

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"kusionstack.io/kusion/pkg/domain/entity"
)

func TestGetResourceQuery(t *testing.T) {
	testcases := []struct {
		name          string
		filter        *entity.ResourceFilter
		expectedQuery string
		expectedArgs  []interface{}
	}{
		{
			name: "filter by organization ID",
			filter: &entity.ResourceFilter{
				OrgID: 42,
			},
			expectedQuery: "organization_id = ?",
			expectedArgs:  []interface{}{"42"},
		},
		{
			name: "filter by project ID and stack ID",
			filter: &entity.ResourceFilter{
				ProjectID: 1,
				StackID:   2,
			},
			expectedQuery: "project_id = ? AND stack_id = ?",
			expectedArgs:  []interface{}{"1", "2"},
		},
		{
			name: "filter by resource plane and resource type",
			filter: &entity.ResourceFilter{
				ResourcePlane: "plane1",
				ResourceType:  "type1",
			},
			expectedQuery: "resource_plane = ? AND resource_type = ?",
			expectedArgs:  []interface{}{"plane1", "type1"},
		},
		{
			name: "filter by all fields",
			filter: &entity.ResourceFilter{
				OrgID:         42,
				ProjectID:     1,
				StackID:       2,
				ResourcePlane: "plane1",
				ResourceType:  "type1",
			},
			expectedQuery: "organization_id = ? AND project_id = ? AND stack_id = ? AND resource_plane = ? AND resource_type = ?",
			expectedArgs:  []interface{}{"42", "1", "2", "plane1", "type1"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			query, args := GetResourceQuery(tc.filter)
			assert.Equal(t, tc.expectedQuery, query)
			assert.Equal(t, tc.expectedArgs, args)
		})
	}
}

func TestGetWorkspaceQuery(t *testing.T) {
	testcases := []struct {
		name          string
		filter        *entity.WorkspaceFilter
		expectedQuery string
		expectedArgs  []interface{}
	}{
		{
			name: "filter by backend ID",
			filter: &entity.WorkspaceFilter{
				BackendID: 42,
			},
			expectedQuery: "backend_id = ?",
			expectedArgs:  []interface{}{"42"},
		},
		{
			name: "filter by name",
			filter: &entity.WorkspaceFilter{
				Name: "example",
			},
			expectedQuery: "name = ?",
			expectedArgs:  []interface{}{"example"},
		},
		// Add more test cases here if needed
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			query, args := GetWorkspaceQuery(tc.filter)
			assert.Equal(t, tc.expectedQuery, query)
			assert.Equal(t, tc.expectedArgs, args)
		})
	}
}

func TestGetStackQuery(t *testing.T) {
	testcases := []struct {
		name          string
		filter        *entity.StackFilter
		expectedQuery string
		expectedArgs  []interface{}
	}{
		{
			name: "filter by organization ID",
			filter: &entity.StackFilter{
				OrgID: 42,
			},
			expectedQuery: "Project.organization_id = ?",
			expectedArgs:  []interface{}{"42"},
		},
		{
			name: "filter by project ID",
			filter: &entity.StackFilter{
				ProjectID: 1,
			},
			expectedQuery: "project_id = ?",
			expectedArgs:  []interface{}{"1"},
		},
		{
			name: "filter by path",
			filter: &entity.StackFilter{
				Path: "example",
			},
			expectedQuery: "Stack.path = ?",
			expectedArgs:  []interface{}{"example"},
		},
		{
			name: "filter by all fields",
			filter: &entity.StackFilter{
				OrgID:     42,
				ProjectID: 1,
				Path:      "example",
			},
			expectedQuery: "Project.organization_id = ? AND project_id = ? AND Stack.path = ?",
			expectedArgs:  []interface{}{"42", "1", "example"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			query, args := GetStackQuery(tc.filter)
			assert.Equal(t, tc.expectedQuery, query)
			assert.Equal(t, tc.expectedArgs, args)
		})
	}
}

func TestGetProjectQuery(t *testing.T) {
	testcases := []struct {
		name          string
		filter        *entity.ProjectFilter
		expectedQuery string
		expectedArgs  []interface{}
	}{
		{
			name: "filter by organization ID",
			filter: &entity.ProjectFilter{
				OrgID: 42,
			},
			expectedQuery: "organization_id = ?",
			expectedArgs:  []interface{}{"42"},
		},
		{
			name: "filter by name",
			filter: &entity.ProjectFilter{
				Name: "example",
			},
			expectedQuery: "name = ?",
			expectedArgs:  []interface{}{"example"},
		},
		// Add more test cases here if needed
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			query, args := GetProjectQuery(tc.filter)
			assert.Equal(t, tc.expectedQuery, query)
			assert.Equal(t, tc.expectedArgs, args)
		})
	}
}

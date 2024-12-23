package persistence

import (
	"testing"
	"time"

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
			expectedQuery: "project.organization_id = ?",
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
			expectedQuery: "stack.path = ?",
			expectedArgs:  []interface{}{"example"},
		},
		{
			name: "filter by all fields",
			filter: &entity.StackFilter{
				OrgID:     42,
				ProjectID: 1,
				Path:      "example",
			},
			expectedQuery: "project.organization_id = ? AND project_id = ? AND stack.path = ?",
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

func TestGetRunQuery(t *testing.T) {
	testcases := []struct {
		name          string
		filter        *entity.RunFilter
		expectedQuery string
		expectedArgs  []interface{}
	}{
		{
			name: "filter by project ID",
			filter: &entity.RunFilter{
				ProjectID: 42,
			},
			expectedQuery: "project.ID = ?",
			expectedArgs:  []interface{}{"42"},
		},
		{
			name: "filter by stack ID and workspace",
			filter: &entity.RunFilter{
				StackID:   1,
				Workspace: "dev",
			},
			expectedQuery: "stack_id = ? AND workspace.name = ?",
			expectedArgs:  []interface{}{uint(1), "dev"},
		},
		{
			name: "filter by type and status",
			filter: &entity.RunFilter{
				Type:   []string{"Preview", "Apply"},
				Status: []string{"Failed", "Succeeded"},
			},
			expectedQuery: "run.type IN (?) AND run.status IN (?)",
			expectedArgs:  []interface{}{[]string{"Preview", "Apply"}, []string{"Failed", "Succeeded"}},
		},
		{
			name: "filter by start time and end time",
			filter: &entity.RunFilter{
				StartTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
				EndTime:   time.Date(2024, 12, 31, 23, 59, 59, 0, time.Local),
			},
			expectedQuery: "run.created_at >= ? AND run.created_at <= ?",
			expectedArgs: []interface{}{
				time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
				time.Date(2024, 12, 31, 23, 59, 59, 0, time.Local),
			},
		},
		{
			name: "filter by all fields",
			filter: &entity.RunFilter{
				ProjectID: 42,
				StackID:   1,
				Workspace: "dev",
				Type:      []string{"Preview"},
				Status:    []string{"Succeeded"},
			},
			expectedQuery: "project.ID = ? AND stack_id = ? AND workspace.name = ? AND run.type IN (?) AND run.status IN (?)",
			expectedArgs:  []interface{}{"42", uint(1), "dev", []string{"Preview"}, []string{"Succeeded"}},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			query, args := GetRunQuery(tc.filter)
			assert.Equal(t, tc.expectedQuery, query)
			assert.Equal(t, tc.expectedArgs, args)
		})
	}
}

func TestGetSourceQuery(t *testing.T) {
	testcases := []struct {
		name          string
		filter        *entity.SourceFilter
		expectedQuery string
		expectedArgs  []interface{}
	}{
		{
			name: "filter by source name",
			filter: &entity.SourceFilter{
				SourceName: "test-source",
			},
			expectedQuery: "source.name LIKE ?",
			expectedArgs:  []interface{}{"%test-source%"},
		},
		{
			name: "filter by source name with special characters",
			filter: &entity.SourceFilter{
				SourceName: "test/source",
			},
			expectedQuery: "source.name LIKE ?",
			expectedArgs:  []interface{}{"%test/source%"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			query, args := GetSourceQuery(tc.filter)
			assert.Equal(t, tc.expectedQuery, query)
			assert.Equal(t, tc.expectedArgs, args)
		})
	}
}

func TestGetModuleQuery(t *testing.T) {
	testcases := []struct {
		name          string
		filter        *entity.ModuleFilter
		expectedQuery string
		expectedArgs  []interface{}
	}{
		{
			name: "filter by module name",
			filter: &entity.ModuleFilter{
				ModuleName: "test-module",
			},
			expectedQuery: "module.name LIKE ?",
			expectedArgs:  []interface{}{"%test-module%"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			query, args := GetModuleQuery(tc.filter)
			assert.Equal(t, tc.expectedQuery, query)
			assert.Equal(t, tc.expectedArgs, args)
		})
	}
}

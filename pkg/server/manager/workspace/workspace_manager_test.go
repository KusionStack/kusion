package workspace

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	backend "kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/domain/constant"
	entity "kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
)

type mockWorkspaceRepository struct {
	mock.Mock
}

func (m *mockWorkspaceRepository) GetByName(ctx context.Context, name string) (*entity.Workspace, error) {
	args := m.Called(ctx, name)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.Workspace), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockWorkspaceRepository) GetByID(ctx context.Context, id uint) (*entity.Workspace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.Workspace), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockWorkspaceRepository) Create(ctx context.Context, workspace *entity.Workspace) error {
	args := m.Called(ctx, workspace)
	return args.Error(0)
}

func (m *mockWorkspaceRepository) Update(ctx context.Context, workspace *entity.Workspace) error {
	args := m.Called(ctx, workspace)
	return args.Error(0)
}

func (m *mockWorkspaceRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockWorkspaceRepository) List(ctx context.Context, filter *entity.WorkspaceFilter, sortOptions *entity.SortOptions) (*entity.WorkspaceListResult, error) {
	args := m.Called(ctx, filter, sortOptions)
	return &entity.WorkspaceListResult{
		Workspaces: args.Get(0).([]*entity.Workspace),
		Total:      len(args.Get(0).([]*entity.Workspace)),
	}, args.Error(1)
}

func (m *mockWorkspaceRepository) Get(ctx context.Context, id uint) (*entity.Workspace, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.Workspace), args.Error(1)
}

type mockBackendRepository struct {
	mock.Mock
}

func (m *mockBackendRepository) GetByName(ctx context.Context, name string) (*entity.Backend, error) {
	args := m.Called(ctx, name)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.Backend), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockBackendRepository) GetByID(ctx context.Context, id uint) (*entity.Backend, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.Backend), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockBackendRepository) Create(ctx context.Context, backend *entity.Backend) error {
	args := m.Called(ctx, backend)
	return args.Error(0)
}

func (m *mockBackendRepository) Update(ctx context.Context, backend *entity.Backend) error {
	args := m.Called(ctx, backend)
	return args.Error(0)
}

func (m *mockBackendRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockBackendRepository) List(ctx context.Context, filter *entity.BackendFilter, sortOptions *entity.SortOptions) (*entity.BackendListResult, error) {
	args := m.Called(ctx, filter, sortOptions)
	return &entity.BackendListResult{
		Backends: args.Get(0).([]*entity.Backend),
		Total:    len(args.Get(0).([]*entity.Backend)),
	}, args.Error(1)
}

func (m *mockBackendRepository) Get(ctx context.Context, id uint) (*entity.Backend, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.Backend), args.Error(1)
}

func TestInvalidNewBackendFromEntity(t *testing.T) {
	// Test cases
	testcases := []struct {
		name           string
		backendEntity  entity.Backend
		expectedResult backend.Backend
		expectedError  error
	}{
		{
			name: "Invalid backend type",
			backendEntity: entity.Backend{
				Name: "invalid name",
				BackendConfig: v1.BackendConfig{
					Type: "invalid",
					// Add required fields for invalid backend configuration
				},
				// Add other fields for the backend entity
			},
			expectedResult: nil,
			expectedError:  fmt.Errorf("invalid type %s of backend %s", "invalid", "invalid name"),
		},
	}

	// Test execution
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := NewBackendFromEntity(tc.backendEntity)
			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestWorkspaceManager_ListWorkspaces(t *testing.T) {
	ctx := context.TODO()
	filter := &entity.WorkspaceFilter{
		// Set your desired filter parameters here
	}
	sortOptions := &entity.SortOptions{
		Field: constant.SortByID,
	}

	// Create a mock workspace repository
	mockRepo := &mockWorkspaceRepository{}
	// Set the expected return value for the List method
	expectedWorkspaces := []*entity.Workspace{
		// Set your expected workspace entities here
	}
	mockRepo.On("List", ctx, filter, sortOptions).Return(expectedWorkspaces, nil)

	// Create a new WorkspaceManager instance with the mock repository
	manager := &WorkspaceManager{
		workspaceRepo: mockRepo,
	}

	// Call the ListWorkspaces method
	workspaces, err := manager.ListWorkspaces(ctx, filter, sortOptions)

	// Assert that the returned stacks match the expected stacks
	if !reflect.DeepEqual(workspaces.Workspaces, expectedWorkspaces) {
		t.Errorf("ListWorkspaces() returned unexpected workspaces.\nExpected: %v\nGot: %v", expectedWorkspaces, workspaces)
	}

	// Assert that no error occurred
	if err != nil {
		t.Errorf("ListWorkspaces() returned an unexpected error: %v", err)
	}

	// Assert that the List method of the mock repository was called with the correct parameters
	mockRepo.AssertCalled(t, "List", ctx, filter, sortOptions)
}

func TestWorkspaceManager_GetWorkspaceByID(t *testing.T) {
	ctx := context.TODO()
	id := uint(1)
	// Create a mock workspace repository
	mockRepo := &mockWorkspaceRepository{}
	// Set the expected return value for the Get method
	expectedWorkspace := &entity.Workspace{
		// Set your expected workspace entity here
	}
	mockRepo.On("Get", ctx, id).Return(expectedWorkspace, nil)
	// Create a new WorkspaceManager instance with the mock repository
	manager := &WorkspaceManager{
		workspaceRepo: mockRepo,
	}
	// Call the GetWorkspaceByID method
	workspace, err := manager.GetWorkspaceByID(ctx, id)
	// Assert that the returned workspace matches the expected workspace
	if !reflect.DeepEqual(workspace, expectedWorkspace) {
		t.Errorf("GetWorkspaceByID() returned unexpected workspace.\nExpected: %v\nGot: %v", expectedWorkspace, workspace)
	}
	// Assert that no error occurred
	if err != nil {
		t.Errorf("GetWorkspaceByID() returned an unexpected error: %v", err)
	}
	// Assert that the Get method of the mock repository was called with the correct parameters
	mockRepo.AssertCalled(t, "Get", ctx, id)
}

func TestWorkspaceManager_DeleteWorkspaceByID(t *testing.T) {
	ctx := context.TODO()
	id := uint(1)
	// Create a mock workspace and backend repository
	mockWorkspaceRepo := &mockWorkspaceRepository{}
	mockBackendRepo := &mockBackendRepository{}

	mockWorkspaceRepo.On("Get", ctx, id).Return(&entity.Workspace{
		ID: id,
		Backend: &entity.Backend{
			ID: 1,
		},
	}, nil)
	mockWorkspaceRepo.On("Delete", ctx, id).Return(nil)
	mockBackendRepo.On("Get", ctx, id).Return(&entity.Backend{
		ID: id,
		BackendConfig: v1.BackendConfig{
			Type: v1.BackendTypeLocal,
		},
	}, nil)
	manager := &WorkspaceManager{
		workspaceRepo: mockWorkspaceRepo,
		backendRepo:   mockBackendRepo,
	}
	err := manager.DeleteWorkspaceByID(ctx, id)
	if err != nil {
		t.Errorf("DeleteWorkspaceByID() returned an unexpected error: %v", err)
	}
}

func TestWorkspaceManager_UpdateWorkspaceByID(t *testing.T) {
	ctx := context.TODO()
	id := uint(1)
	// Create a mock workspace and backend repository
	mockWorkspaceRepo := &mockWorkspaceRepository{}
	mockBackendRepo := &mockBackendRepository{}
	// Set the expected return value for the Get method
	expectedBackend := &entity.Backend{
		ID: 1,
		BackendConfig: v1.BackendConfig{
			Type: v1.BackendTypeLocal,
		},
	}

	updatedWorkspace := &entity.Workspace{
		ID:      id,
		Name:    "dev",
		Backend: expectedBackend,
	}
	mockWorkspaceRepo.On("Get", ctx, id).Return(updatedWorkspace, nil)
	mockWorkspaceRepo.On("Update", ctx, updatedWorkspace).Return(nil)
	mockBackendRepo.On("Get", ctx, id).Return(expectedBackend, nil)
	// Create a new WorkspaceManager instance with the mock repository
	manager := &WorkspaceManager{
		workspaceRepo: mockWorkspaceRepo,
		backendRepo:   mockBackendRepo,
	}
	// Call the UpdateWorkspaceByID method
	workspace, err := manager.UpdateWorkspaceByID(ctx, id, request.UpdateWorkspaceRequest{
		Name: "dev",
	})
	// Assert that the returned workspace matches the expected workspace
	if !reflect.DeepEqual(workspace, updatedWorkspace) {
		t.Errorf("UpdateWorkspaceByID() returned unexpected workspace.\nExpected: %v\nGot: %v", updatedWorkspace, workspace)
	}
	// Assert that no error occurred
	if err != nil {
		t.Errorf("UpdateWorkspaceByID() returned an unexpected error: %v", err)
	}
	// Assert that the Update method of the mock repository was called with the correct parameters
	mockWorkspaceRepo.AssertCalled(t, "Update", ctx, updatedWorkspace)
}

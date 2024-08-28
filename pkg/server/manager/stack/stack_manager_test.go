package stack

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
)

type mockStackRepository struct {
	mock.Mock
}

func (m *mockStackRepository) GetByName(ctx context.Context, name string) (*entity.Stack, error) {
	args := m.Called(ctx, name)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.Stack), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockStackRepository) GetByID(ctx context.Context, id uint) (*entity.Stack, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.Stack), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockStackRepository) Create(ctx context.Context, workspace *entity.Stack) error {
	args := m.Called(ctx, workspace)
	return args.Error(0)
}

func (m *mockStackRepository) Update(ctx context.Context, workspace *entity.Stack) error {
	args := m.Called(ctx, workspace)
	return args.Error(0)
}

func (m *mockStackRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockStackRepository) List(ctx context.Context, filter *entity.StackFilter) ([]*entity.Stack, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*entity.Stack), args.Error(1)
}

func (m *mockStackRepository) Get(ctx context.Context, id uint) (*entity.Stack, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.Stack), args.Error(1)
}

func TestStackManager_ListStacks(t *testing.T) {
	ctx := context.TODO()
	filter := &entity.StackFilter{
		// Set your desired filter parameters here
	}

	// Create a mock stack repository
	mockRepo := &mockStackRepository{}
	// Set the expected return value for the List method
	expectedStacks := []*entity.Stack{
		// Set your expected stack entities here
	}
	mockRepo.On("List", ctx, filter).Return(expectedStacks, nil)

	// Create a new StackManager instance with the mock repository
	manager := &StackManager{
		stackRepo: mockRepo,
	}

	// Call the ListStacks method
	stacks, err := manager.ListStacks(ctx, filter)

	// Assert that the returned stacks match the expected stacks
	if !reflect.DeepEqual(stacks, expectedStacks) {
		t.Errorf("ListStacks() returned unexpected stacks.\nExpected: %v\nGot: %v", expectedStacks, stacks)
	}

	// Assert that no error occurred
	if err != nil {
		t.Errorf("ListStacks() returned an unexpected error: %v", err)
	}

	// Assert that the List method of the mock repository was called with the correct parameters
	mockRepo.AssertCalled(t, "List", ctx, filter)
}

func TestStackManager_GetStackByID(t *testing.T) {
	ctx := context.TODO()
	id := uint(1)
	// Create a mock stack repository
	mockRepo := &mockStackRepository{}
	// Set the expected return value for the Get method
	expectedStack := &entity.Stack{
		// Set your expected stack entity here
	}
	mockRepo.On("Get", ctx, id).Return(expectedStack, nil)
	// Create a new StackManager instance with the mock repository
	manager := &StackManager{
		stackRepo: mockRepo,
	}
	// Call the GetStackByID method
	stack, err := manager.GetStackByID(ctx, id)
	// Assert that the returned stack matches the expected stack
	if !reflect.DeepEqual(stack, expectedStack) {
		t.Errorf("GetStackByID() returned unexpected stack.\nExpected: %v\nGot: %v", expectedStack, stack)
	}
	// Assert that no error occurred
	if err != nil {
		t.Errorf("GetStackByID() returned an unexpected error: %v", err)
	}
	// Assert that the Get method of the mock repository was called with the correct parameters
	mockRepo.AssertCalled(t, "Get", ctx, id)
}

func TestStackManager_DeleteStackByID(t *testing.T) {
	ctx := context.TODO()
	id := uint(1)
	// Create a mock stack repository
	mockRepo := &mockStackRepository{}
	// Set the expected return value for the Delete method
	mockRepo.On("Delete", ctx, id).Return(nil)
	// Create a new StackManager instance with the mock repository
	manager := &StackManager{
		stackRepo: mockRepo,
	}
	// Call the DeleteStackByID method
	err := manager.DeleteStackByID(ctx, id)
	// Assert that no error occurred
	if err != nil {
		t.Errorf("DeleteStackByID() returned an unexpected error: %v", err)
	}
	// Assert that the Delete method of the mock repository was called with the correct parameters
	mockRepo.AssertCalled(t, "Delete", ctx, id)
}

func TestStackManager_UpdateStackByID(t *testing.T) {
	ctx := context.TODO()
	id := uint(1)
	requestPayload := request.UpdateStackRequest{
		CreateStackRequest: request.CreateStackRequest{
			DesiredVersion: "v1.0.0",
		},
	}
	mockStackRepo := &mockStackRepository{}
	mockProjectRepo := &mockProjectRepository{}
	expectedProject := &entity.Project{
		Name: "test-project",
	}
	mockProjectRepo.On("Get", ctx, requestPayload.ProjectID).Return(expectedProject, nil)
	expectedStack := &entity.Stack{
		Name:           "test-stack",
		Project:        expectedProject,
		Path:           "test-stack-path",
		SyncState:      "UnSynced",
		DesiredVersion: "v1.0.0",
	}
	mockStackRepo.On("Get", ctx, id).Return(expectedStack, nil)
	// Set the expected return value for the Update method
	mockStackRepo.On("Update", ctx, mock.Anything).Return(nil)
	// Create a new StackManager instance with the mock repository
	manager := &StackManager{
		stackRepo:   mockStackRepo,
		projectRepo: mockProjectRepo,
	}
	// Call the UpdateStackByID method
	stack, err := manager.UpdateStackByID(ctx, id, requestPayload)
	if err != nil {
		t.Errorf("UpdateStackByID() returned an unexpected error: %v", err)
	}
	assert.Equal(t, expectedStack.DesiredVersion, stack.DesiredVersion)
}

func TestStackManager_CreateStack(t *testing.T) {
	ctx := context.TODO()
	expectedProject := &entity.Project{
		Name: "test-project",
		Path: "test-project-path",
	}
	expectedStack := &entity.Stack{
		Name:      "test-stack",
		Project:   expectedProject,
		Path:      "test-stack-path",
		SyncState: "UnSynced",
	}
	mockStackRepo := &mockStackRepository{}
	mockProjectRepo := &mockProjectRepository{}
	t.Run("CreateStack", func(t *testing.T) {
		requestPayload := request.CreateStackRequest{
			Name:      "test-stack",
			ProjectID: 1,
			Path:      "test-stack-path",
		}
		mockProjectRepo.On("Get", ctx, requestPayload.ProjectID).Return(expectedProject, nil)
		mockStackRepo.On("Create", ctx, mock.Anything).Return(nil)
		manager := &StackManager{
			stackRepo:   mockStackRepo,
			projectRepo: mockProjectRepo,
		}
		stack, err := manager.CreateStack(ctx, requestPayload)
		if err != nil {
			t.Errorf("CreateStack() returned an unexpected error: %v", err)
		}
		assert.Equal(t, expectedStack.Name, stack.Name)
		assert.Equal(t, expectedStack.Path, stack.Path)
	})
	t.Run("CreateStackWithProjectName", func(t *testing.T) {
		requestPayload := request.CreateStackRequest{
			Name:        "test-stack",
			ProjectName: "test-project",
		}
		mockProjectRepo.On("GetByName", ctx, requestPayload.ProjectName).Return(expectedProject, nil)
		mockStackRepo.On("Create", ctx, mock.Anything).Return(nil)
		manager := &StackManager{
			stackRepo:   mockStackRepo,
			projectRepo: mockProjectRepo,
		}
		stack, err := manager.CreateStack(ctx, requestPayload)
		if err != nil {
			t.Errorf("CreateStack() returned an unexpected error: %v", err)
		}
		assert.Equal(t, expectedStack.Name, stack.Name)
		assert.Equal(t, "test-project-path/test-stack", stack.Path)
	})
}

type mockProjectRepository struct {
	mock.Mock
}

func (m *mockProjectRepository) GetByName(ctx context.Context, name string) (*entity.Project, error) {
	args := m.Called(ctx, name)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.Project), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockProjectRepository) GetByID(ctx context.Context, id uint) (*entity.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.Project), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockProjectRepository) Create(ctx context.Context, workspace *entity.Project) error {
	args := m.Called(ctx, workspace)
	return args.Error(0)
}

func (m *mockProjectRepository) Update(ctx context.Context, workspace *entity.Project) error {
	args := m.Called(ctx, workspace)
	return args.Error(0)
}

func (m *mockProjectRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockProjectRepository) List(ctx context.Context, filter *entity.ProjectFilter) ([]*entity.Project, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*entity.Project), args.Error(1)
}

func (m *mockProjectRepository) Get(ctx context.Context, id uint) (*entity.Project, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.Project), args.Error(1)
}

package project

import (
	"context"
	"net/url"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
)

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

func (m *mockProjectRepository) List(ctx context.Context, filter *entity.ProjectFilter) (*entity.ProjectListResult, error) {
	args := m.Called(ctx, filter)
	return &entity.ProjectListResult{
		Projects: args.Get(0).([]*entity.Project),
		Total:    len(args.Get(0).([]*entity.Project)),
	}, args.Error(1)
}

func (m *mockProjectRepository) Get(ctx context.Context, id uint) (*entity.Project, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.Project), args.Error(1)
}

type mockOrganizationRepository struct {
	mock.Mock
}

func (m *mockOrganizationRepository) GetByName(ctx context.Context, name string) (*entity.Organization, error) {
	args := m.Called(ctx, name)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.Organization), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockOrganizationRepository) GetByID(ctx context.Context, id uint) (*entity.Organization, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.Organization), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockOrganizationRepository) Create(ctx context.Context, workspace *entity.Organization) error {
	args := m.Called(ctx, workspace)
	return args.Error(0)
}

func (m *mockOrganizationRepository) Update(ctx context.Context, workspace *entity.Organization) error {
	args := m.Called(ctx, workspace)
	return args.Error(0)
}

func (m *mockOrganizationRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockOrganizationRepository) List(ctx context.Context, filter *entity.OrganizationFilter) (*entity.OrganizationListResult, error) {
	args := m.Called(ctx)
	return &entity.OrganizationListResult{
		Organizations: args.Get(0).([]*entity.Organization),
		Total:         len(args.Get(0).([]*entity.Organization)),
	}, args.Error(1)
}

func (m *mockOrganizationRepository) Get(ctx context.Context, id uint) (*entity.Organization, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.Organization), args.Error(1)
}

type mockSourceRepository struct {
	mock.Mock
}

func (m *mockSourceRepository) GetByName(ctx context.Context, name string) (*entity.Source, error) {
	args := m.Called(ctx, name)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.Source), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockSourceRepository) GetByID(ctx context.Context, id uint) (*entity.Source, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.Source), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockSourceRepository) Create(ctx context.Context, workspace *entity.Source) error {
	args := m.Called(ctx, workspace)
	return args.Error(0)
}

func (m *mockSourceRepository) Update(ctx context.Context, workspace *entity.Source) error {
	args := m.Called(ctx, workspace)
	return args.Error(0)
}

func (m *mockSourceRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockSourceRepository) List(ctx context.Context, filter *entity.SourceFilter) (*entity.SourceListResult, error) {
	args := m.Called(ctx)
	return &entity.SourceListResult{
		Sources: args.Get(0).([]*entity.Source),
		Total:   len(args.Get(0).([]*entity.Source)),
	}, args.Error(1)
}

func (m *mockSourceRepository) Get(ctx context.Context, id uint) (*entity.Source, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.Source), args.Error(1)
}

func (m *mockSourceRepository) GetByRemote(ctx context.Context, remote string) (*entity.Source, error) {
	args := m.Called(ctx, remote)
	return args.Get(0).(*entity.Source), args.Error(1)
}

func TestProjectManager_ListProjects(t *testing.T) {
	ctx := context.TODO()
	filter := &entity.ProjectFilter{}
	mockRepo := &mockProjectRepository{}
	expectedProjects := []*entity.Project{
		{
			ID:   1,
			Name: "Project 1",
		},
	}
	mockRepo.On("List", ctx, filter).Return(expectedProjects, nil)
	manager := &ProjectManager{
		projectRepo: mockRepo,
	}
	projects, err := manager.ListProjects(ctx, filter)
	if !reflect.DeepEqual(projects.Projects, expectedProjects) {
		t.Errorf("ListProjects() returned unexpected projects.\nExpected: %v\nGot: %v", expectedProjects, projects)
	}
	if err != nil {
		t.Errorf("ListProjects() returned an unexpected error: %v", err)
	}
	mockRepo.AssertCalled(t, "List", ctx, filter)
}

func TestProjectManager_GetProjectByID(t *testing.T) {
	ctx := context.TODO()
	id := uint(1)
	mockRepo := &mockProjectRepository{}
	expectedProject := &entity.Project{
		ID:   1,
		Name: "Project 1",
	}
	mockRepo.On("Get", ctx, id).Return(expectedProject, nil)
	manager := &ProjectManager{
		projectRepo: mockRepo,
	}
	project, err := manager.GetProjectByID(ctx, id)
	if !reflect.DeepEqual(project, expectedProject) {
		t.Errorf("GetProjectByID() returned unexpected project.\nExpected: %v\nGot: %v", expectedProject, project)
	}
	if err != nil {
		t.Errorf("GetProjectByID() returned an unexpected error: %v", err)
	}
	mockRepo.AssertCalled(t, "Get", ctx, id)
}

func TestProjectManager_DeleteProjectByID(t *testing.T) {
	ctx := context.TODO()
	id := uint(1)
	mockRepo := &mockProjectRepository{}
	mockRepo.On("Delete", ctx, id).Return(nil)
	manager := &ProjectManager{
		projectRepo: mockRepo,
	}
	err := manager.DeleteProjectByID(ctx, id)
	if err != nil {
		t.Errorf("DeleteProjectByID() returned an unexpected error: %v", err)
	}
	mockRepo.AssertCalled(t, "Delete", ctx, id)
}

func TestProjectManager_UpdateProjectByID(t *testing.T) {
	ctx := context.TODO()
	id := uint(1)
	requestPayload := request.UpdateProjectRequest{
		SourceID:       2,
		OrganizationID: 3,
		Labels:         []string{"label1", "label2"},
	}
	mockRepo := &mockProjectRepository{}
	mockSourceRepo := &mockSourceRepository{}
	mockOrganizationRepo := &mockOrganizationRepository{}
	expectedProject := &entity.Project{
		ID:   1,
		Name: "Project 1",
	}
	mockRepo.On("Get", ctx, id).Return(expectedProject, nil)
	expectedSource := &entity.Source{
		ID: 2,
		Remote: &url.URL{
			Scheme: "https",
			Host:   "github.com",
		},
	}
	mockSourceRepo.On("Get", ctx, requestPayload.SourceID).Return(expectedSource, nil)
	expectedOrganization := &entity.Organization{
		ID:   3,
		Name: "Organization 1",
	}
	mockOrganizationRepo.On("Get", ctx, requestPayload.OrganizationID).Return(expectedOrganization, nil)
	mockRepo.On("Update", ctx, expectedProject).Return(nil)
	manager := &ProjectManager{
		projectRepo:      mockRepo,
		sourceRepo:       mockSourceRepo,
		organizationRepo: mockOrganizationRepo,
		defaultSource: entity.Source{
			Remote: &url.URL{
				Scheme: "https",
				Host:   "github.com",
			},
		},
	}
	updatedProject, err := manager.UpdateProjectByID(ctx, id, requestPayload)
	if err != nil {
		t.Errorf("UpdateProjectByID() returned an unexpected error: %v", err)
	}
	if !reflect.DeepEqual(updatedProject, expectedProject) {
		t.Errorf("UpdateProjectByID() returned unexpected project.\nExpected: %v\nGot: %v", expectedProject, updatedProject)
	}
	mockRepo.AssertCalled(t, "Get", ctx, id)
	mockSourceRepo.AssertCalled(t, "Get", ctx, requestPayload.SourceID)
	mockOrganizationRepo.AssertCalled(t, "Get", ctx, requestPayload.OrganizationID)
	mockRepo.AssertCalled(t, "Update", ctx, expectedProject)
}

func TestProjectManager_CreateProject(t *testing.T) {
	ctx := context.TODO()
	mockRepo := &mockProjectRepository{}
	mockSourceRepo := &mockSourceRepository{}
	mockOrganizationRepo := &mockOrganizationRepository{}
	manager := &ProjectManager{
		projectRepo:      mockRepo,
		sourceRepo:       mockSourceRepo,
		organizationRepo: mockOrganizationRepo,
		defaultSource: entity.Source{
			Remote: &url.URL{
				Scheme: "https",
				Host:   "github.com",
			},
		},
	}
	expectedSource := &entity.Source{
		ID: 1,
		Remote: &url.URL{
			Scheme: "https",
			Host:   "github.com",
		},
	}
	t.Run("CreateProject", func(t *testing.T) {
		requestPayload := request.CreateProjectRequest{
			Name:           "Project 1",
			SourceID:       2,
			OrganizationID: 3,
			Labels:         []string{"label1", "label2"},
		}
		mockSourceRepo.On("Get", ctx, requestPayload.SourceID).Return(expectedSource, nil)
		expectedOrganization := &entity.Organization{
			ID:   3,
			Name: "Organization 1",
		}
		mockOrganizationRepo.On("Get", ctx, requestPayload.OrganizationID).Return(expectedOrganization, nil)
		mockRepo.On("Create", ctx, mock.Anything).Return(nil)
		createdProject, err := manager.CreateProject(ctx, requestPayload)
		if err != nil {
			t.Errorf("CreateProject() returned an unexpected error: %v", err)
		}
		if createdProject.Name != requestPayload.Name {
			t.Errorf("CreateProject() returned unexpected project.\nExpected: %v\nGot: %v", requestPayload.Name, createdProject.Name)
		}
		mockSourceRepo.AssertCalled(t, "Get", ctx, requestPayload.SourceID)
		mockOrganizationRepo.AssertCalled(t, "Get", ctx, requestPayload.OrganizationID)
	})
	t.Run("CreateProjectWithDefaultSource", func(t *testing.T) {
		requestPayload := request.CreateProjectRequest{
			Name:           "Project 1",
			OrganizationID: 3,
			Labels:         []string{"label1", "label2"},
		}
		expectedOrganization := &entity.Organization{
			ID:   3,
			Name: "Organization 1",
		}
		mockOrganizationRepo.On("Get", ctx, requestPayload.OrganizationID).Return(expectedOrganization, nil)
		mockSourceRepo.On("GetByRemote", ctx, "https://github.com").Return(expectedSource, nil)
		mockRepo.On("Create", ctx, mock.Anything).Return(nil)
		createdProject, err := manager.CreateProject(ctx, requestPayload)
		if err != nil {
			t.Errorf("CreateProject() returned an unexpected error: %v", err)
		}
		if createdProject.Name != requestPayload.Name {
			t.Errorf("CreateProject() returned unexpected project.\nExpected: %v\nGot: %v", requestPayload.Name, createdProject.Name)
		}
		mockOrganizationRepo.AssertCalled(t, "Get", ctx, requestPayload.OrganizationID)
	})
	t.Run("CreateProjectWithDomain", func(t *testing.T) {
		requestPayload := request.CreateProjectRequest{
			Name:   "Project 1",
			Domain: "new-org",
			Labels: []string{"label1", "label2"},
		}
		expectedOrganization := &entity.Organization{
			ID:   3,
			Name: "Organization 1",
		}
		mockOrganizationRepo.On("GetByName", ctx, requestPayload.Domain).Return(expectedOrganization, nil)
		mockSourceRepo.On("GetByRemote", ctx, "https://github.com").Return(expectedSource, nil)
		mockRepo.On("Create", ctx, mock.Anything).Return(nil)
		createdProject, err := manager.CreateProject(ctx, requestPayload)
		if err != nil {
			t.Errorf("CreateProject() returned an unexpected error: %v", err)
		}
		if createdProject.Name != requestPayload.Name {
			t.Errorf("CreateProject() returned unexpected project.\nExpected: %v\nGot: %v", requestPayload.Name, createdProject.Name)
		}
		mockOrganizationRepo.AssertCalled(t, "GetByName", ctx, requestPayload.Domain)
	})
}

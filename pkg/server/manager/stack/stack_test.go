package stack

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform/tfops"
	"kusionstack.io/kusion/pkg/infra/persistence"
)

func TestBuildOptions(t *testing.T) {
	dryrun := true
	maxConcurrent := 10

	options := BuildOptions(dryrun, maxConcurrent)

	assert.NotNil(t, options)
	assert.Equal(t, dryrun, options.DryRun)
	assert.Equal(t, maxConcurrent, options.MaxConcurrent)
}

func TestProcessChanges(t *testing.T) {
	changes := &models.Changes{
		ChangeOrder: &models.ChangeOrder{
			ChangeSteps: map[string]*models.ChangeStep{
				"step1": {
					ID:     "id",
					Action: models.Update,
					From:   "old value",
					To:     "new value",
				},
			},
		},
	}

	format := "json"
	detail := true

	result, err := ProcessChanges(context.Background(), nil, changes, format, detail)

	assert.NoError(t, err)
	assert.Equal(t, changes, result)
}

func TestGetBackendFromWorkspaceName(t *testing.T) {
	m := &StackManager{
		workspaceRepo: &mockWorkspaceRepository{},
		defaultBackend: entity.Backend{
			ID:   1,
			Name: "default",
			BackendConfig: v1.BackendConfig{
				Type: "local",
			},
		},
	}

	ctx := context.Background()

	t.Run("DefaultWorkspace", func(t *testing.T) {
		workspaceName := constant.DefaultWorkspace

		backend, err := m.getBackendFromWorkspaceName(ctx, workspaceName)

		assert.NoError(t, err)
		assert.NotNil(t, backend)
	})

	t.Run("NonDefaultWorkspace", func(t *testing.T) {
		workspaceName := "my-workspace"

		workspaceEntity := &entity.Workspace{
			ID:   2,
			Name: workspaceName,
			Backend: &entity.Backend{
				ID:   3,
				Name: "remote",
				BackendConfig: v1.BackendConfig{
					Type: "local",
				},
			},
		}

		m.workspaceRepo.(*mockWorkspaceRepository).On("GetByName", ctx, workspaceName).Return(workspaceEntity, nil)

		backend, err := m.getBackendFromWorkspaceName(ctx, workspaceName)

		assert.NoError(t, err)
		assert.NotNil(t, backend)
	})

	t.Run("NonDefaultWorkspace_NotFound", func(t *testing.T) {
		workspaceName := "non-existing-workspace"

		m.workspaceRepo.(*mockWorkspaceRepository).On("GetByName", ctx, workspaceName).Return(nil, gorm.ErrRecordNotFound)

		backend, err := m.getBackendFromWorkspaceName(ctx, workspaceName)

		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		assert.Nil(t, backend)
	})
}

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

func TestGetStackProjectAndBackend(t *testing.T) {
	m := &StackManager{
		workspaceRepo: &mockWorkspaceRepository{},
		defaultBackend: entity.Backend{
			ID:   1,
			Name: "default",
			BackendConfig: v1.BackendConfig{
				Type: "local",
			},
		},
	}

	t.Run("DefaultWorkspace", func(t *testing.T) {
		workspaceName := constant.DefaultWorkspace
		// Create a mock stack entity
		stackEntity := &entity.Stack{
			Project: &entity.Project{},
		}

		// Call the getStackProjectAndBackend function
		project, stack, backend, err := m.getStackProjectAndBackend(context.Background(), stackEntity, workspaceName)

		// Assert that the returned values are correct
		assert.NotNil(t, project)
		assert.NotNil(t, stack)
		assert.NotNil(t, backend)
		assert.NoError(t, err)
	})
}

func TestGetDefaultBackend(t *testing.T) {
	m := &StackManager{
		defaultBackend: entity.Backend{
			ID:   1,
			Name: "default",
			BackendConfig: v1.BackendConfig{
				Type: "local",
			},
		},
	}

	backend, err := m.getDefaultBackend()

	assert.NoError(t, err)
	assert.NotNil(t, backend)
}

func TestBuildValidStackPath(t *testing.T) {
	projectEntity := &entity.Project{
		Path: "myproject",
	}

	testcases := []struct {
		name           string
		requestPayload request.CreateStackRequest
		expectedPath   string
		expectedValid  bool
	}{
		{
			name: "valid stack path without cloud",
			requestPayload: request.CreateStackRequest{
				Name: "mystack",
			},
			expectedPath:  "myproject/mystack",
			expectedValid: true,
		},
		{
			name: "invalid stack path",
			requestPayload: request.CreateStackRequest{
				Name: "my_!@$!@#%stack",
			},
			expectedPath:  "myproject/my_!@$!@#%stack",
			expectedValid: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path, valid := buildValidStackPath(tc.requestPayload, projectEntity)
			assert.Equal(t, tc.expectedPath, path)
			assert.Equal(t, tc.expectedValid, valid)
		})
	}
}

func TestBuildStackFilterAndSortOptions(t *testing.T) {
	m := &StackManager{
		projectRepo: &mockProjectRepository{},
	}
	ctx := context.Background()

	t.Run("Valid IDs", func(t *testing.T) {
		query := &url.Values{}
		query.Add("orgID", "123")
		query.Add("projectID", "456")
		query.Add("projectName", "")
		query.Add("env", "")
		filter, sortOptions, err := m.BuildStackFilterAndSortOptions(ctx, query)
		assert.NoError(t, err)
		assert.Equal(t, uint(123), filter.OrgID)
		assert.Equal(t, uint(456), filter.ProjectID)
		assert.Equal(t, constant.SortByID, sortOptions.Field)
	})

	// Test case 1: Valid organization ID and project ID
	t.Run("Invalid organization ID", func(t *testing.T) {
		query := &url.Values{}
		query.Add("orgID", "abc")
		query.Add("projectID", "456")
		query.Add("projectName", "")
		query.Add("env", "")
		_, _, err := m.BuildStackFilterAndSortOptions(ctx, query)
		assert.Error(t, err)
		assert.Equal(t, constant.ErrInvalidOrganizationID, err)
	})

	t.Run("Invalid project ID", func(t *testing.T) {
		query := &url.Values{}
		query.Add("orgID", "")
		query.Add("projectID", "def")
		query.Add("projectName", "")
		query.Add("env", "")
		_, _, err := m.BuildStackFilterAndSortOptions(ctx, query)
		assert.Error(t, err)
		assert.Equal(t, constant.ErrInvalidProjectID, err)
	})

	t.Run("Valid project Name", func(t *testing.T) {
		projectName := "projectName"
		expectedProject := &entity.Project{
			ID:   1,
			Name: projectName,
		}
		m.projectRepo.(*mockProjectRepository).On("GetByName", ctx, projectName).Return(expectedProject, nil)

		query := &url.Values{}
		query.Add("orgID", "")
		query.Add("env", "")
		query.Add("projectName", projectName)
		query.Add("sortBy", constant.SortByCreateTimestamp)
		query.Add("ascending", "true")
		filter, sortOptions, err := m.BuildStackFilterAndSortOptions(ctx, query)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), filter.ProjectID)
		assert.Equal(t, "created_at", sortOptions.Field)
		assert.Equal(t, true, sortOptions.Ascending)
	})
}

func TestImportTerraformResourceID(t *testing.T) {
	m := &StackManager{
		defaultBackend: entity.Backend{
			ID:   1,
			Name: "default",
			BackendConfig: v1.BackendConfig{
				Type: "local",
			},
		},
	}

	sp := &v1.Spec{
		Resources: []v1.Resource{
			{
				Type: v1.Terraform,
				ID:   "tf-resource1",
				Extensions: map[string]interface{}{
					"provider": "aws",
				},
			},
			{
				Type: v1.Terraform,
				ID:   "tf-resource2",
				Extensions: map[string]interface{}{
					"provider": "azure",
				},
			},
			{
				Type: v1.Kubernetes,
				ID:   "k8s-resource1",
			},
		},
	}

	importedResources := map[string]string{
		"tf-resource1": "arn:aws:resource1",
		"tf-resource2": "azure-resource2",
	}

	m.ImportTerraformResourceID(context.Background(), sp, importedResources)

	expected := &v1.Spec{
		Resources: []v1.Resource{
			{
				Type: v1.Terraform,
				ID:   "tf-resource1",
				Extensions: map[string]interface{}{
					"provider": "aws",
				},
			},
			{
				Type: v1.Terraform,
				ID:   "tf-resource2",
				Extensions: map[string]interface{}{
					"provider": "azure",
				},
			},
			{
				Type: v1.Kubernetes,
				ID:   "k8s-resource1",
			},
		},
	}

	expected.Resources[0].Extensions[tfops.ImportIDKey] = "arn:aws:resource1"
	expected.Resources[1].Extensions[tfops.ImportIDKey] = "azure-resource2"

	assert.Equal(t, expected, sp)
}

func TestConvertV1ResourceToEntity(t *testing.T) {
	t.Run("Kubernetes Resource", func(t *testing.T) {
		resource := &v1.Resource{
			ID:   "apps/v1:Deployment:my-namespace:my-deployment",
			Type: v1.Kubernetes,
			Extensions: map[string]interface{}{
				"provider": "Kubernetes",
			},
			Attributes: map[string]interface{}{
				"replicas": 3,
			},
		}

		expectedResource := &entity.Resource{
			KusionResourceID: "apps/v1:Deployment:my-namespace:my-deployment",
			IAMResourceID:    "",
			CloudResourceID:  "",
			ResourcePlane:    "Kubernetes",
			ResourceType:     "apps/v1/Deployment",
			ResourceName:     "my-namespace/my-deployment",
		}

		result, err := convertV1ResourceToEntity(resource)
		assert.NoError(t, err)
		assert.Equal(t, expectedResource, result)
	})
	t.Run("Kubernetes Resource without namespace", func(t *testing.T) {
		resource := &v1.Resource{
			ID:   "v1:Namespace:my-namespace",
			Type: v1.Kubernetes,
			Extensions: map[string]interface{}{
				"provider": "Kubernetes",
			},
			Attributes: map[string]interface{}{
				"name": "my-namespace",
			},
		}

		expectedResource := &entity.Resource{
			KusionResourceID: "v1:Namespace:my-namespace",
			IAMResourceID:    "",
			CloudResourceID:  "",
			ResourcePlane:    "Kubernetes",
			ResourceType:     "v1/Namespace",
			ResourceName:     "my-namespace",
		}

		result, err := convertV1ResourceToEntity(resource)
		assert.NoError(t, err)
		assert.Equal(t, expectedResource, result)
	})
	t.Run("AWS Resource", func(t *testing.T) {
		resource := &v1.Resource{
			ID:   "aws:aws:ec2_instance:my-instance",
			Type: v1.Terraform,
			Extensions: map[string]interface{}{
				"provider": "aws",
			},
			Attributes: map[string]interface{}{
				"instance_type": "t2.micro",
				"arn":           "arn:aws:ec2_instance:my-instance",
			},
		}

		expectedResource := &entity.Resource{
			KusionResourceID: "aws:aws:ec2_instance:my-instance",
			IAMResourceID:    "",
			CloudResourceID:  "arn:aws:ec2_instance:my-instance",
			ResourcePlane:    "aws",
			ResourceType:     "ec2_instance",
			ResourceName:     "my-instance",
			Provider:         "aws",
		}
		result, err := convertV1ResourceToEntity(resource)
		assert.NoError(t, err)
		assert.Equal(t, expectedResource, result)
	})
	t.Run("AliCloud Resource", func(t *testing.T) {
		resource := &v1.Resource{
			ID:   "alicloud:alicloud:ecs_instance:my-instance",
			Type: v1.Terraform,
			Extensions: map[string]interface{}{
				"provider": "alicloud",
			},
			Attributes: map[string]interface{}{
				"instance_type": "ecs.t5-lc1m2.small",
				"id":            "arn:alicloud:ecs_instance:my-instance",
			},
		}

		expectedResource := &entity.Resource{
			KusionResourceID: "alicloud:alicloud:ecs_instance:my-instance",
			IAMResourceID:    "",
			CloudResourceID:  "arn:alicloud:ecs_instance:my-instance",
			ResourcePlane:    "alicloud",
			ResourceType:     "ecs_instance",
			ResourceName:     "my-instance",
			Provider:         "alicloud",
		}
		result, err := convertV1ResourceToEntity(resource)
		assert.NoError(t, err)
		assert.Equal(t, expectedResource, result)
	})
	t.Run("Azure Resource", func(t *testing.T) {
		resource := &v1.Resource{
			ID:   "azure:azure:vm:my-vm",
			Type: v1.Terraform,
			Extensions: map[string]interface{}{
				"provider": "azure",
			},
			Attributes: map[string]interface{}{
				"size": "Standard_B1s",
				"id":   "azure:vm:my-vm",
			},
		}

		expectedResource := &entity.Resource{
			KusionResourceID: "azure:azure:vm:my-vm",
			IAMResourceID:    "",
			CloudResourceID:  "azure:vm:my-vm",
			ResourcePlane:    "azure",
			ResourceType:     "vm",
			ResourceName:     "my-vm",
			Provider:         "azure",
		}
		result, err := convertV1ResourceToEntity(resource)
		assert.NoError(t, err)
		assert.Equal(t, expectedResource, result)
	})
	t.Run("GCP Resource", func(t *testing.T) {
		resource := &v1.Resource{
			ID:   "google:google:compute_instance:my-instance",
			Type: v1.Terraform,
			Extensions: map[string]interface{}{
				"provider": "google",
			},
			Attributes: map[string]interface{}{
				"machine_type": "n1-standard-1",
				"id":           "google:compute_instance:my-instance",
			},
		}

		expectedResource := &entity.Resource{
			KusionResourceID: "google:google:compute_instance:my-instance",
			IAMResourceID:    "",
			CloudResourceID:  "google:compute_instance:my-instance",
			ResourcePlane:    "google",
			ResourceType:     "compute_instance",
			ResourceName:     "my-instance",
			Provider:         "google",
		}
		result, err := convertV1ResourceToEntity(resource)
		assert.NoError(t, err)
		assert.Equal(t, expectedResource, result)
	})
}

func TestNewStackManager(t *testing.T) {
	fakeGDB := &gorm.DB{}
	stackRepo := &mockStackRepository{}
	projectRepo := &mockProjectRepository{}
	workspaceRepo := &mockWorkspaceRepository{}
	resourceRepo := persistence.NewResourceRepository(fakeGDB)
	runRepo := persistence.NewRunRepository(fakeGDB)
	defaultBackend := entity.Backend{}
	maxConcurrent := 10

	manager := NewStackManager(stackRepo, projectRepo, workspaceRepo, resourceRepo, runRepo, defaultBackend, maxConcurrent)

	assert.NotNil(t, manager)
	assert.Equal(t, stackRepo, manager.stackRepo)
	assert.Equal(t, projectRepo, manager.projectRepo)
	assert.Equal(t, workspaceRepo, manager.workspaceRepo)
	assert.Equal(t, resourceRepo, manager.resourceRepo)
	assert.Equal(t, defaultBackend, manager.defaultBackend)
	assert.Equal(t, maxConcurrent, manager.maxConcurrent)
}

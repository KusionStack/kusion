package repository

import (
	"context"

	"kusionstack.io/kusion/pkg/domain/entity"
)

// OrganizationRepository is an interface that defines the repository operations
// for organizations. It follows the principles of domain-driven design (DDD).
type OrganizationRepository interface {
	// Create creates a new organization.
	Create(ctx context.Context, organization *entity.Organization) error
	// Delete deletes a organization by its ID.
	Delete(ctx context.Context, id uint) error
	// Update updates an existing organization.
	Update(ctx context.Context, organization *entity.Organization) error
	// Get retrieves a organization by its ID.
	Get(ctx context.Context, id uint) (*entity.Organization, error)
	// GetByName retrieves a organization by its name.
	GetByName(ctx context.Context, name string) (*entity.Organization, error)
	// List retrieves all existing organizations.
	List(ctx context.Context) ([]*entity.Organization, error)
}

// ProjectRepository is an interface that defines the repository operations
// for projects. It follows the principles of domain-driven design (DDD).
type ProjectRepository interface {
	// Create creates a new project.
	Create(ctx context.Context, project *entity.Project) error
	// Delete deletes a project by its ID.
	Delete(ctx context.Context, id uint) error
	// Update updates an existing project.
	Update(ctx context.Context, project *entity.Project) error
	// Get retrieves a project by its ID.
	Get(ctx context.Context, id uint) (*entity.Project, error)
	// GetByName retrieves a project by its name.
	GetByName(ctx context.Context, name string) (*entity.Project, error)
	// List retrieves all existing projects.
	List(ctx context.Context, filter *entity.ProjectFilter) ([]*entity.Project, error)
}

// StackRepository is an interface that defines the repository operations
// for stacks. It follows the principles of domain-driven design (DDD).
type StackRepository interface {
	// Create creates a new stack.
	Create(ctx context.Context, stack *entity.Stack) error
	// Delete deletes a stack by its ID.
	Delete(ctx context.Context, id uint) error
	// Update updates an existing stack.
	Update(ctx context.Context, stack *entity.Stack) error
	// Get retrieves a stack by its ID.
	Get(ctx context.Context, id uint) (*entity.Stack, error)
	// List retrieves all existing stacks.
	List(ctx context.Context, filter *entity.StackFilter) ([]*entity.Stack, error)
}

// SourceRepository is an interface that defines the repository operations
// for sources. It follows the principles of domain-driven design (DDD).
type SourceRepository interface {
	// Get retrieves a source by its ID.
	Get(ctx context.Context, id uint) (*entity.Source, error)
	// GetByRemote retrieves a source by its remote.
	GetByRemote(ctx context.Context, remote string) (*entity.Source, error)
	// List retrieves all existing sources.
	List(ctx context.Context) ([]*entity.Source, error)
	// Create creates a new source.
	Create(ctx context.Context, source *entity.Source) error
	// Delete deletes a stack by its ID.
	Delete(ctx context.Context, id uint) error
	// Update updates an existing stack.
	Update(ctx context.Context, stack *entity.Source) error
}

// WorkspaceRepository is an interface that defines the repository operations
// for workspaces. It follows the principles of domain-driven design (DDD).
type WorkspaceRepository interface {
	// Create creates a new workspace.
	Create(ctx context.Context, workspace *entity.Workspace) error
	// Delete deletes a workspace by its ID.
	Delete(ctx context.Context, id uint) error
	// Update updates an existing workspace.
	Update(ctx context.Context, workspace *entity.Workspace) error
	// Get retrieves a workspace by its ID.
	Get(ctx context.Context, id uint) (*entity.Workspace, error)
	// GetByName retrieves a workspace by its name.
	GetByName(ctx context.Context, name string) (*entity.Workspace, error)
	// List retrieves all existing workspace.
	List(ctx context.Context, filter *entity.WorkspaceFilter) ([]*entity.Workspace, error)
}

// BackendRepository is an interface that defines the repository operations
// for backends. It follows the principles of domain-driven design (DDD).
type BackendRepository interface {
	// Create creates a new backend.
	Create(ctx context.Context, backend *entity.Backend) error
	// Delete deletes a backend by its ID.
	Delete(ctx context.Context, id uint) error
	// Update updates an existing backend.
	Update(ctx context.Context, backend *entity.Backend) error
	// Get retrieves a backend by its ID.
	Get(ctx context.Context, id uint) (*entity.Backend, error)
	// List retrieves all existing backend.
	List(ctx context.Context) ([]*entity.Backend, error)
}

// ResourceRepository is an interface that defines the repository operations
// for resources. It follows the principles of domain-driven design (DDD).
type ResourceRepository interface {
	// Create creates a new resource.
	Create(ctx context.Context, resource []*entity.Resource) error
	// Delete deletes a resource by its ID.
	Delete(ctx context.Context, id uint) error
	// Batch deletes a list of resources
	BatchDelete(ctx context.Context, resource []*entity.Resource) error
	// Update updates an existing resource.
	Update(ctx context.Context, resource *entity.Resource) error
	// Get retrieves a resource by its ID.
	Get(ctx context.Context, id uint) (*entity.Resource, error)
	// GetByKusionResourceURN retrieves a resource by its Kusion resource URN.
	GetByKusionResourceURN(ctx context.Context, urn string) (*entity.Resource, error)
	// List retrieves all existing resource.
	List(ctx context.Context, filter *entity.ResourceFilter) ([]*entity.Resource, error)
}

// ModuleRepository is an interface that defines the repository operations
// for Kusion Modules. It follows the principles of domain-driven design (DDD).
type ModuleRepository interface {
	// Create creates a new module.
	Create(ctx context.Context, module *entity.Module) error
	// Delete deletes a module by its name.
	Delete(ctx context.Context, name string) error
	// Update updates an existing module.
	Update(ctx context.Context, module *entity.Module) error
	// Get retrieves a module by its name.
	Get(ctx context.Context, name string) (*entity.Module, error)
	// List retrives all the existing modules.
	List(ctx context.Context) ([]*entity.Module, error)
}

// RunRepository is an interface that defines the repository operations
// for runs. It follows the principles of domain-driven design (DDD).
type RunRepository interface {
	// Create creates a new run.
	Create(ctx context.Context, run *entity.Run) error
	// Delete deletes a run by its ID.
	Delete(ctx context.Context, id uint) error
	// Update updates an existing run.
	Update(ctx context.Context, run *entity.Run) error
	// Get retrieves a run by its ID.
	Get(ctx context.Context, id uint) (*entity.Run, error)
	// List retrieves all existing run.
	List(ctx context.Context, filter *entity.RunFilter) (*entity.RunListResult, error)
}

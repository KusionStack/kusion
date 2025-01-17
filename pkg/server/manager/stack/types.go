package stack

import (
	"errors"

	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/repository"
	cache "kusionstack.io/kusion/pkg/server/util/cache"
)

const (
	Stdout      = "stdout"
	NoDiffFound = "All resources are reconciled. No diff found"
)

var (
	ErrGettingNonExistingStack                   = errors.New("the stack does not exist")
	ErrUpdatingNonExistingStack                  = errors.New("the stack to update does not exist")
	ErrInvalidStackID                            = errors.New("the stack ID should be a uuid")
	ErrCanOnlyUpdateConfigItemInNonStandardStack = errors.New("can only update config item in non-standard stack")
	ErrGettingNonExistingStateForStack           = errors.New("can not find State in this stack")
	ErrNoManagedResourceToDestroy                = errors.New("no managed resources to destroy")
	ErrDryrunApply                               = errors.New("dryrun-mode is enabled, no resources will be applied")
	ErrDryrunDestroy                             = errors.New("dryrun-mode is enabled, no resources will be destroyed")
	ErrStackInOperation                          = errors.New("the stack is being operated by another request. Please wait until it is completed")
	ErrStackNotPreviewedYet                      = errors.New("the stack has not been previewed yet. Please generate and preview the stack first")
	ErrInvalidRunID                              = errors.New("the run ID should be a uuid")
	ErrInvalidWatchTimeout                       = errors.New("watchTimeout should be a number")
	ErrWorkspaceEmpty                            = errors.New("workspace should not be empty in query")
	ErrRunRequestBodyEmpty                       = errors.New("run request body should not be empty")
	ErrRunCrashed                                = errors.New("run crashed")
)

type StackManager struct {
	stackRepo      repository.StackRepository
	projectRepo    repository.ProjectRepository
	workspaceRepo  repository.WorkspaceRepository
	resourceRepo   repository.ResourceRepository
	runRepo        repository.RunRepository
	defaultBackend entity.Backend
	maxConcurrent  int
	repoCache      *cache.Cache[uint, *StackCache]
}

type StackCache struct {
	LocalDirOnDisk string
	StackPath      string
}

type StackRequestParams struct {
	StackID       uint
	Workspace     string
	Format        string
	Operator      string
	ExecuteParams StackExecuteParams
}

type StackExecuteParams struct {
	Detail              bool
	Dryrun              bool
	SpecID              string
	Force               bool
	ImportResources     bool
	NoCache             bool
	Unlock              bool
	Watch               bool
	WatchTimeoutSeconds int
}

type RunRequestParams struct {
	RunID uint
}

func NewStackManager(stackRepo repository.StackRepository,
	projectRepo repository.ProjectRepository,
	workspaceRepo repository.WorkspaceRepository,
	resourceRepo repository.ResourceRepository,
	runRepo repository.RunRepository,
	defaultBackend entity.Backend,
	maxConcurrent int,
) *StackManager {
	return &StackManager{
		stackRepo:      stackRepo,
		projectRepo:    projectRepo,
		workspaceRepo:  workspaceRepo,
		resourceRepo:   resourceRepo,
		runRepo:        runRepo,
		defaultBackend: defaultBackend,
		maxConcurrent:  maxConcurrent,
		repoCache:      cache.NewCache[uint, *StackCache](constant.RepoCacheTTL),
	}
}

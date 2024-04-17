package stack

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/repository"
	"kusionstack.io/kusion/pkg/domain/request"

	engineapi "kusionstack.io/kusion/pkg/engine/api"
	"kusionstack.io/kusion/pkg/engine/operation/models"

	sourceapi "kusionstack.io/kusion/pkg/engine/api/source"
	"kusionstack.io/kusion/pkg/server/handler"
	"kusionstack.io/kusion/pkg/server/util"
)

func NewStackManager(stackRepo repository.StackRepository, projectRepo repository.ProjectRepository, workspaceRepo repository.WorkspaceRepository) *StackManager {
	return &StackManager{
		stackRepo:     stackRepo,
		projectRepo:   projectRepo,
		workspaceRepo: workspaceRepo,
	}
}

func (m *StackManager) GenerateStack(ctx context.Context, id uint, workspaceName string) (*v1.Spec, error) {
	logger := util.GetLogger(ctx)
	logger.Info("Starting generating spec in StackManager ...")

	// Generate a stack
	stackEntity, err := m.stackRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingStack
		}
		return nil, err
	}

	// Get project by id
	project, err := stackEntity.Project.ConvertToCore()
	if err != nil {
		return nil, err
	}

	// Get stack by id
	stack, err := stackEntity.ConvertToCore()
	if err != nil {
		return nil, err
	}

	// Get workspace configurations from backend
	wsBackend, err := m.getBackendFromWorkspaceName(ctx, workspaceName)
	if err != nil {
		return nil, err
	}
	wsStorage, err := wsBackend.WorkspaceStorage()
	if err != nil {
		return nil, err
	}
	ws, err := wsStorage.Get(workspaceName)
	if err != nil {
		return nil, err
	}

	// Build API inputs
	// get project to get source and workdir
	projectEntity, err := handler.GetProjectByID(ctx, m.projectRepo, stackEntity.Project.ID)
	if err != nil {
		return nil, err
	}

	directory, workDir, err := GetWorkDirFromSource(ctx, stackEntity, projectEntity)
	logger.Info("workDir derived", "workDir", workDir)
	logger.Info("directory derived", "directory", directory)

	stack.Path = workDir
	if err != nil {
		return nil, err
	}
	// intentOptions, _ := buildOptions(workDir, kpmParam, false)
	// Cleanup
	defer sourceapi.Cleanup(ctx, directory)

	// Generate spec
	return engineapi.GenerateSpecWithSpinner(project, stack, ws, true)
}

func (m *StackManager) PreviewStack(ctx context.Context, id uint, workspaceName string) (*models.Changes, error) {
	logger := util.GetLogger(ctx)
	logger.Info("Starting previewing stack in StackManager ...")
	_, changes, _, err := m.previewHelper(ctx, id, workspaceName)
	return changes, err
}

func (m *StackManager) ApplyStack(ctx context.Context, id uint, workspaceName, format string, detail, dryrun bool, w http.ResponseWriter) error {
	logger := util.GetLogger(ctx)
	logger.Info("Starting applying stack in StackManager ...")

	// Get the stack entity by id
	stackEntity, err := m.stackRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGettingNonExistingStack
		}
		return err
	}

	// Preview a stack
	sp, changes, stateStorage, err := m.previewHelper(ctx, id, workspaceName)
	if err != nil {
		return err
	}

	_, err = ProcessChanges(ctx, w, changes, format, detail)
	if err != nil {
		return err
	}

	// if dry run, print the hint
	if dryrun {
		logger.Info("NOTE: Currently running in the --dry-run mode, the above configuration does not really take effect")
		return ErrDryrunDestroy
	}

	logger.Info("Dryrun set to false. Start applying diffs ...")
	executeOptions := BuildOptions(dryrun)
	if err = engineapi.Apply(executeOptions, stateStorage, sp, changes, os.Stdout); err != nil {
		return err
	}

	// Update LastSyncTimestamp to current time and set stack syncState to synced
	stackEntity.LastSyncTimestamp = time.Now()
	stackEntity.SyncState = constant.StackStateSynced

	// Update stack with repository
	err = m.stackRepo.Update(ctx, stackEntity)
	if err != nil {
		return err
	}

	return nil
}

func (m *StackManager) DestroyStack(ctx context.Context, id uint, workspaceName string, detail, dryrun bool, w http.ResponseWriter) error {
	logger := util.GetLogger(ctx)
	logger.Info("Starting applying stack in StackManager ...")

	// Get the stack entity by id
	stackEntity, err := m.stackRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGettingNonExistingStack
		}
		return err
	}

	// Get project by id
	project, err := stackEntity.Project.ConvertToCore()
	if err != nil {
		return err
	}

	// Get stack by id
	stack, err := stackEntity.ConvertToCore()
	if err != nil {
		return err
	}

	stateBackend, err := m.getBackendFromWorkspaceName(ctx, workspaceName)
	if err != nil {
		return err
	}

	// Build API inputs
	// get project to get source and workdir
	projectEntity, err := handler.GetProjectByID(ctx, m.projectRepo, stackEntity.Project.ID)
	if err != nil {
		return err
	}

	directory, workDir, err := GetWorkDirFromSource(ctx, stackEntity, projectEntity)
	if err != nil {
		return err
	}
	destroyOptions := BuildOptions(dryrun)
	stack.Path = workDir

	// Cleanup
	defer sourceapi.Cleanup(ctx, directory)

	// Compute state storage
	stateStorage := stateBackend.StateStorage(project.Name, workspaceName)
	logger.Info("Remote state storage found", "Remote", stateStorage)

	priorState, err := stateStorage.Get()
	if err != nil || priorState == nil {
		logger.Info("can't find state", "project", project.Name, "stack", stack.Name, "workspace", workspaceName)
		return ErrGettingNonExistingStateForStack
	}
	destroyResources := priorState.Resources

	if destroyResources == nil || len(priorState.Resources) == 0 {
		return ErrNoManagedResourceToDestroy
	}

	// compute changes for preview
	i := &v1.Spec{Resources: destroyResources}
	changes, err := engineapi.DestroyPreview(destroyOptions, i, project, stack, stateStorage)
	if err != nil {
		return err
	}

	// Summary preview table
	changes.Summary(w, true)
	// detail detection
	if detail {
		changes.OutputDiff("all")
	}

	// if dryrun, print the hint
	if dryrun {
		logger.Info("Dry-run mode enabled, the above resources will be destroyed if dryrun is set to false")
		return ErrDryrunDestroy
	}

	// Destroy
	logger.Info("Start destroying resources......")
	if err = engineapi.Destroy(destroyOptions, i, changes, stateStorage); err != nil {
		return err
	}
	return nil
}

func (m *StackManager) ListStacks(ctx context.Context) ([]*entity.Stack, error) {
	stackEntities, err := m.stackRepo.List(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingStack
		}
		return nil, err
	}
	return stackEntities, nil
}

func (m *StackManager) GetStackByID(ctx context.Context, id uint) (*entity.Stack, error) {
	existingEntity, err := m.stackRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingStack
		}
		return nil, err
	}
	return existingEntity, nil
}

func (m *StackManager) DeleteStackByID(ctx context.Context, id uint) error {
	err := m.stackRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGettingNonExistingStack
		}
		return err
	}
	return nil
}

func (m *StackManager) UpdateStackByID(ctx context.Context, id uint, requestPayload request.UpdateStackRequest) (*entity.Stack, error) {
	// Convert request payload to domain model
	var requestEntity entity.Stack
	if err := copier.Copy(&requestEntity, &requestPayload); err != nil {
		return nil, err
	}

	// Get project by id
	projectEntity, err := handler.GetProjectByID(ctx, m.projectRepo, requestPayload.ProjectID)
	if err != nil {
		return nil, err
	}
	requestEntity.Project = projectEntity

	// Get the existing stack by id
	updatedEntity, err := m.stackRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUpdatingNonExistingStack
		}
		return nil, err
	}

	// Overwrite non-zero values in request entity to existing entity
	copier.CopyWithOption(updatedEntity, requestEntity, copier.Option{IgnoreEmpty: true})

	// Update stack with repository
	err = m.stackRepo.Update(ctx, updatedEntity)
	if err != nil {
		return nil, err
	}
	return updatedEntity, nil
}

func (m *StackManager) CreateStack(ctx context.Context, requestPayload request.CreateStackRequest) (*entity.Stack, error) {
	// Convert request payload to domain model
	var createdEntity entity.Stack
	if err := copier.Copy(&createdEntity, &requestPayload); err != nil {
		return nil, err
	}
	// The default state is UnSynced
	createdEntity.SyncState = constant.StackStateUnSynced
	createdEntity.CreationTimestamp = time.Now()
	createdEntity.UpdateTimestamp = time.Now()
	createdEntity.LastSyncTimestamp = time.Unix(0, 0) // default to none

	// Get project by id
	projectEntity, err := handler.GetProjectByID(ctx, m.projectRepo, requestPayload.ProjectID)
	if err != nil {
		return nil, err
	}
	createdEntity.Project = projectEntity

	// Create stack with repository
	err = m.stackRepo.Create(ctx, &createdEntity)
	if err != nil {
		return nil, err
	}
	return &createdEntity, nil
}

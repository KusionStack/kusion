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
	"kusionstack.io/kusion/pkg/engine/release"

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
	opts, stackBackend, project, stack, ws, err := m.metaHelper(ctx, id, workspaceName)
	if err != nil {
		return nil, err
	}

	// Generate spec
	sp, err := engineapi.GenerateSpecWithSpinner(project, stack, ws, true)
	if err != nil {
		return nil, err
	}
	// return immediately if no resource found in stack
	// todo: if there is no resource, should still do diff job; for now, if output is json format, there is no hint
	if sp == nil || len(sp.Resources) == 0 {
		logger.Info("No resource change found in this stack...")
		return nil, nil
	}

	// Preview
	releaseStorage, err := stackBackend.ReleaseStorage(project.Name, ws.Name)
	if err != nil {
		return nil, err
	}
	state, err := release.GetLatestState(releaseStorage)
	if err != nil {
		return nil, err
	}
	if state == nil {
		state = &v1.State{}
	}
	changes, err := engineapi.Preview(opts, releaseStorage, sp, state, project, stack)
	return changes, err
}

func (m *StackManager) ApplyStack(ctx context.Context, id uint, workspaceName, format string, detail, dryrun bool, w http.ResponseWriter) (err error) {
	logger := util.GetLogger(ctx)
	logger.Info("Starting applying stack in StackManager ...")

	var storage release.Storage
	var rel *v1.Release
	releaseCreated := false
	defer func() {
		if !releaseCreated {
			return
		}
		if err != nil {
			rel.Phase = v1.ReleasePhaseFailed
			_ = release.UpdateApplyRelease(storage, rel, dryrun)
		} else {
			rel.Phase = v1.ReleasePhaseSucceeded
			err = release.UpdateApplyRelease(storage, rel, dryrun)
		}
	}()

	// create release
	opts, stackBackend, project, stack, ws, err := m.metaHelper(ctx, id, workspaceName)
	if err != nil {
		return err
	}
	storage, err = stackBackend.ReleaseStorage(project.Name, ws.Name)
	if err != nil {
		return
	}
	rel, err = release.NewApplyRelease(storage, project.Name, stack.Name, ws.Name)
	if err != nil {
		return
	}
	if !dryrun {
		if err = storage.Create(rel); err != nil {
			return
		}
		releaseCreated = true
	}

	// Get the stack entity by id
	stackEntity, err := m.stackRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGettingNonExistingStack
		}
		return
	}

	// generate spec
	sp, err := engineapi.GenerateSpecWithSpinner(project, stack, ws, true)
	if err != nil {
		return
	}
	// return immediately if no resource found in stack
	// todo: if there is no resource, should still do diff job; for now, if output is json format, there is no hint
	if sp == nil || len(sp.Resources) == 0 {
		logger.Info("No resource change found in this stack...")
		return nil
	}

	// update release phase to previewing
	rel.Spec = sp
	rel.Phase = v1.ReleasePhasePreviewing
	if err = release.UpdateApplyRelease(storage, rel, dryrun); err != nil {
		return
	}
	// compute changes for preview
	changes, err := engineapi.Preview(opts, storage, rel.Spec, rel.State, project, stack)
	if err != nil {
		return
	}
	_, err = ProcessChanges(ctx, w, changes, format, detail)
	if err != nil {
		return
	}

	// if dry run, print the hint
	if dryrun {
		logger.Info("NOTE: Currently running in the --dry-run mode, the above configuration does not really take effect")
		return ErrDryrunDestroy
	}

	rel.Phase = v1.ReleasePhaseApplying
	if err = release.UpdateApplyRelease(storage, rel, dryrun); err != nil {
		return
	}
	logger.Info("Dryrun set to false. Start applying diffs ...")
	executeOptions := BuildOptions(dryrun)
	var upRel *v1.Release
	upRel, err = engineapi.Apply(executeOptions, storage, rel, changes, os.Stdout)
	if err != nil {
		return
	}
	rel = upRel

	// Update LastSyncTimestamp to current time and set stack syncState to synced
	stackEntity.LastSyncTimestamp = time.Now()
	stackEntity.SyncState = constant.StackStateSynced

	// Update stack with repository
	err = m.stackRepo.Update(ctx, stackEntity)
	if err != nil {
		return
	}

	return nil
}

func (m *StackManager) DestroyStack(ctx context.Context, id uint, workspaceName string, detail, dryrun bool, w http.ResponseWriter) (err error) {
	logger := util.GetLogger(ctx)
	logger.Info("Starting applying stack in StackManager ...")

	// update release to succeeded or failed
	var storage release.Storage
	var rel *v1.Release
	releaseCreated := false
	defer func() {
		if !releaseCreated {
			return
		}
		if err != nil {
			rel.Phase = v1.ReleasePhaseFailed
			_ = release.UpdateDestroyRelease(storage, rel)
		} else {
			rel.Phase = v1.ReleasePhaseSucceeded
			err = release.UpdateDestroyRelease(storage, rel)
		}
	}()

	// create release
	_, stackBackend, project, stack, ws, err := m.metaHelper(ctx, id, workspaceName)
	if err != nil {
		return err
	}
	storage, err = stackBackend.ReleaseStorage(project.Name, ws.Name)
	if err != nil {
		return
	}
	rel, err = release.CreateDestroyRelease(storage, project.Name, stack.Name, ws.Name)
	if err != nil {
		return
	}
	if len(rel.Spec.Resources) == 0 {
		return ErrNoManagedResourceToDestroy
	}
	releaseCreated = true

	// compute changes for preview
	changes, err := engineapi.DestroyPreview(rel.Spec, rel.State, project, stack, storage)
	if err != nil {
		return
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

	// update release phase to destroying
	rel.Phase = v1.ReleasePhaseDestroying
	if err = release.UpdateDestroyRelease(storage, rel); err != nil {
		return
	}
	// Destroy
	logger.Info("Start destroying resources......")
	var upRel *v1.Release
	upRel, err = engineapi.Destroy(rel, changes, storage)
	if err != nil {
		return
	}
	rel = upRel
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

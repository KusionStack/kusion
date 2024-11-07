package stack

import (
	"context"
	"errors"
	"net/http"
	"os"
	"sync"
	"time"

	"gorm.io/gorm"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/engine/release"

	engineapi "kusionstack.io/kusion/pkg/engine/api"
	"kusionstack.io/kusion/pkg/engine/operation/models"

	sourceapi "kusionstack.io/kusion/pkg/engine/api/source"
	appmiddleware "kusionstack.io/kusion/pkg/server/middleware"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

func (m *StackManager) GenerateSpec(ctx context.Context, params *StackRequestParams) (string, *apiv1.Spec, error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Starting generating spec in StackManager ...")

	// Get the stack entity and return error if stack ID is not found
	stackEntity, err := m.stackRepo.Get(ctx, params.StackID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, ErrGettingNonExistingStack
		}
		return "", nil, err
	}

	// Ensure the state is updated properly
	defer func() {
		if err != nil {
			stackEntity.SyncState = constant.StackStateGenerateFailed
		} else {
			stackEntity.SyncState = constant.StackStateGenerated
		}
		m.stackRepo.Update(ctx, stackEntity)
	}()

	// If the stack is being generated/previewed/applied/destroyed by another request, return an error
	// TODO: This is a temporary solution to prevent multiple requests from generating the same stack and cause concurrency issues
	// To override this, pass in force == true
	if stackEntity.StackInOperation() && !params.ExecuteParams.Force {
		err = ErrStackInOperation
		return "", nil, err
	}

	// Set stack sync state to generating
	stackEntity.SyncState = constant.StackStateGenerating
	err = m.stackRepo.Update(ctx, stackEntity)
	if err != nil {
		return "", nil, err
	}

	// Otherwise, generate spec from stack entity using the default generator
	project, stack, wsBackend, err := m.getStackProjectAndBackend(ctx, stackEntity, params.Workspace)
	wsStorage, err := wsBackend.WorkspaceStorage()
	if err != nil {
		return "", nil, err
	}
	ws, err := wsStorage.Get(params.Workspace)
	if err != nil {
		return "", nil, err
	}

	// Build API inputs
	// get project to get source and workdir
	projectEntity, err := m.projectRepo.Get(ctx, stackEntity.Project.ID)
	if err != nil {
		return "", nil, err
	}

	directory, workDir, err := GetWorkDirFromSource(ctx, stackEntity, projectEntity)
	logger.Info("workDir derived", "workDir", workDir)
	logger.Info("directory derived", "directory", directory)

	stack.Path = workDir
	if err != nil {
		return "", nil, err
	}

	stackEntity.SyncState = constant.StackStateGenerated
	err = m.stackRepo.Update(ctx, stackEntity)
	if err != nil {
		return "", nil, err
	}
	// Cleanup
	defer sourceapi.Cleanup(ctx, directory)

	// Generate spec
	sp, err := engineapi.GenerateSpecWithSpinner(project, stack, ws, true)
	return "", sp, err
}

func (m *StackManager) PreviewStack(ctx context.Context, params *StackRequestParams, requestPayload request.StackImportRequest) (*models.Changes, error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Starting previewing stack in StackManager ...")

	// Get the stack entity by id
	stackEntity, err := m.stackRepo.Get(ctx, params.StackID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingStack
		}
		return nil, err
	}

	defer func() {
		if err != nil {
			stackEntity.SyncState = constant.StackStatePreviewFailed
			m.stackRepo.Update(ctx, stackEntity)
		} else {
			stackEntity.SyncState = constant.StackStatePreviewed
			if params.ExecuteParams.SpecID != "" {
				stackEntity.LastPreviewedRevision = params.ExecuteParams.SpecID
			} else {
				stackEntity.LastPreviewedRevision = stackEntity.LastGeneratedRevision
			}
			m.stackRepo.Update(ctx, stackEntity)
		}
	}()

	// If the stack is being generated/previewed/applied/destroyed by another request, return an error
	// TODO: This is a temporary solution to prevent multiple requests from previewing the same stack and cause concurrency issues
	// To override this, pass in force == true
	if stackEntity.StackInOperation() && !params.ExecuteParams.Force {
		return nil, ErrStackInOperation
	}

	// Set stack sync state to previewing
	stackEntity.SyncState = constant.StackStatePreviewing
	err = m.stackRepo.Update(ctx, stackEntity)
	if err != nil {
		return nil, err
	}

	var sp *apiv1.Spec
	executeOptions := BuildOptions(false, m.maxConcurrent)
	project, stack, stateBackend, err := m.getStackProjectAndBackend(ctx, stackEntity, params.Workspace)
	if err != nil {
		return nil, err
	}
	releasePath := getReleasePath(stackEntity.Path, "default")
	releaseStorage, err := stateBackend.StateStorageWithPath(releasePath)
	if err != nil {
		return nil, err
	}
	logger.Info("State storage found with path", "releasePath", releasePath)

	// Get workspace configurations from backend
	wsStorage, err := stateBackend.WorkspaceStorage()
	if err != nil {
		return nil, err
	}
	ws, err := wsStorage.Get(params.Workspace)
	if err != nil {
		return nil, err
	}

	var directory, workDir string
	// TODO: remove this. This is only for testing purposes
	m.repoCache.Set(18, &StackCache{LocalDirOnDisk: "/tmp/kcp-kusion-3803080093", StackPath: "/tmp/kcp-kusion-3803080093/simpleservice4/dev"})
	if stackCache, exists := m.repoCache.Get(stackEntity.ID); exists {
		// If found in repoCache, use the cached workDir and directory
		logger.Info("Stack found in cache. Using cache...")
		workDir = stackCache.StackPath
		stack.Path = workDir
		// This is temporarily commented out to speed up development
		// directory = stackCache.LocalDirOnDisk
	} else {
		// If not found in repoCache, checkout workdir
		logger.Info("Stack not found in cache. Pulling repo and set cache...")
		directory, workDir, err = GetWorkDirFromSource(ctx, stackEntity, stackEntity.Project)
		if err != nil {
			return nil, err
		}
		stack.Path = workDir
		sc := &StackCache{
			LocalDirOnDisk: directory,
			StackPath:      workDir,
		}
		m.repoCache.Set(stackEntity.ID, sc)
	}

	// Cleanup
	defer func() {
		// This is temporarily commented out to speed up development
		// sourceapi.Cleanup(ctx, directory)
	}()

	// Generate spec using default generator
	sp, err = engineapi.GenerateSpecWithSpinner(project, stack, ws, true)
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
	state, err := release.GetLatestState(releaseStorage)
	if err != nil {
		return nil, err
	}
	if state == nil {
		state = &apiv1.State{}
	}
	stack.Path = tempPath(stackEntity.Path)

	// Set context from workspace to spec
	if ws != nil && len(ws.Context) > 0 {
		sp.Context = ws.Context
	}

	// Set import details if importResources is set to true
	if params.ExecuteParams.ImportResources && len(requestPayload.ImportedResources) > 0 {
		m.ImportTerraformResourceID(ctx, sp, requestPayload.ImportedResources)
	}
	logger.Info("Final Spec is: ", "spec", sp)

	changes, err := engineapi.Preview(executeOptions, releaseStorage, sp, state, project, stack)
	return changes, err
}

func (m *StackManager) ApplyStack(ctx context.Context, params *StackRequestParams, requestPayload request.StackImportRequest) error {
	logger := logutil.GetLogger(ctx)
	logger.Info("Starting applying stack in StackManager ...")
	_, stackBackend, project, _, ws, err := m.metaHelper(ctx, params.StackID, params.Workspace)
	if err != nil {
		return err
	}

	// Get the stack entity by id
	stackEntity, err := m.stackRepo.Get(ctx, params.StackID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGettingNonExistingStack
		}
		return err
	}

	specID := ""
	// If specID is explicitly specified by the caller, use the spec with the specID
	if params.ExecuteParams.SpecID != "" {
		specID = params.ExecuteParams.SpecID
		logger.Info("SpecID explicitly set. Using the specified version", "SpecID", specID)
	} else {
		specID = stackEntity.LastPreviewedRevision
		logger.Info("SpecID not explicitly set. Using last previewed version", "SpecID", stackEntity.LastPreviewedRevision)
	}

	var storage release.Storage
	rel := &apiv1.Release{}
	relLock := &sync.Mutex{}
	releaseCreated := false
	// Ensure the state is updated properly
	defer func() {
		if !releaseCreated {
			return
		}
		if err != nil {
			stackEntity.SyncState = constant.StackStateApplyFailed
			release.UpdateReleasePhase(rel, apiv1.ReleasePhaseFailed, relLock)
			_ = release.UpdateApplyRelease(storage, rel, params.ExecuteParams.Dryrun, relLock)
		} else {
			release.UpdateReleasePhase(rel, apiv1.ReleasePhaseSucceeded, relLock)
			err = release.UpdateApplyRelease(storage, rel, params.ExecuteParams.Dryrun, relLock)
			// Update LastSyncTimestamp to current time and set stack syncState to synced
			if !params.ExecuteParams.Dryrun {
				stackEntity.SyncState = constant.StackStateSynced
				stackEntity.LastAppliedTimestamp = time.Now()
				stackEntity.LastAppliedRevision = specID
			}
		}
		m.stackRepo.Update(ctx, stackEntity)
	}()

	// If the stack is being generated/previewed/applied/destroyed by another request, return an error
	// TODO: This is a temporary solution to prevent multiple requests from applying the same stack and cause concurrency issues
	// To override this, pass in force == true
	if stackEntity.StackInOperation() && !params.ExecuteParams.Force {
		return ErrStackInOperation
	}
	// Temporarily commented out
	// if stackEntity.LastPreviewedRevision == "" || stackEntity.SyncState != constant.StackStatePreviewed {
	// if stackEntity.LastPreviewedRevision == "" {
	// 	// This indicates the stack has not been generated and previewed before
	// 	// We will not allow this to continue until it has been properly previewed
	// 	return ErrStackNotPreviewedYet
	// }

	// Set stack sync state to applying
	stackEntity.SyncState = constant.StackStateApplying
	err = m.stackRepo.Update(ctx, stackEntity)
	if err != nil {
		return err
	}

	// create release
	releasePath := getReleasePath(stackEntity.Path, "default")
	storage, err = stackBackend.StateStorageWithPath(releasePath)
	if err != nil {
		return err
	}
	logger.Info("State storage found with path", "releasePath", releasePath)
	if err != nil {
		return err
	}
	priorState, err := release.GetLatestState(storage)
	if err != nil {
		return err
	}
	if priorState == nil {
		priorState = &apiv1.State{}
	}
	rel, err = release.NewApplyRelease(storage, project.Name, stackEntity.Name, ws.Name)
	if err != nil {
		return err
	}

	if !params.ExecuteParams.Dryrun {
		if err = storage.Create(rel); err != nil {
			return err
		}
		releaseCreated = true
	}

	var sp *apiv1.Spec
	var changes *models.Changes
	project, stack, stateBackend, err := m.getStackProjectAndBackend(ctx, stackEntity, params.Workspace)
	executeOptions := BuildOptions(params.ExecuteParams.Dryrun, m.maxConcurrent)

	logger.Info("Previewing using the default generator ...")
	// Checkout workdir
	directory, workDir, err := GetWorkDirFromSource(ctx, stackEntity, stackEntity.Project)
	if err != nil {
		return err
	}
	stack.Path = workDir

	// Cleanup
	defer sourceapi.Cleanup(ctx, directory)

	// Generate spec using default generator
	sp, err = engineapi.GenerateSpecWithSpinner(project, stack, ws, true)
	if err != nil {
		return err
	}

	// return immediately if no resource found in stack
	// todo: if there is no resource, should still do diff job; for now, if output is json format, there is no hint
	if sp == nil || len(sp.Resources) == 0 {
		logger.Info("No resource change found in this stack...")
		return nil
	}

	// update release phase to previewing
	rel.Spec = sp
	release.UpdateReleasePhase(rel, apiv1.ReleasePhasePreviewing, relLock)
	if err = release.UpdateApplyRelease(storage, rel, params.ExecuteParams.Dryrun, relLock); err != nil {
		return err
	}

	// if dry run, print the hint
	if params.ExecuteParams.Dryrun {
		logger.Info("NOTE: Currently running in the --dry-run mode, the above configuration does not really take effect")
		err = ErrDryrunApply
		return err
	}

	logger.Info("State backend found", "stateBackend", stateBackend)
	stack.Path = tempPath(stackEntity.Path)

	// Set context from workspace to spec
	if ws != nil && len(ws.Context) > 0 {
		sp.Context = ws.Context
		// Set x-kusion-trace in spec context
		sp.Context["x-kusion-trace"] = appmiddleware.GetTraceID(ctx)
		sp.Context["x-kusion-spec-id"] = specID
	}

	// Set import details if importResources is set to true
	if params.ExecuteParams.ImportResources && len(requestPayload.ImportedResources) > 0 {
		m.ImportTerraformResourceID(ctx, sp, requestPayload.ImportedResources)
	}

	// Calculate change steps
	changes, err = engineapi.Preview(executeOptions, storage, sp, priorState, project, stack)
	if err != nil {
		return err
	}

	logger.Info("Dryrun set to false. Start applying diffs ...")
	release.UpdateReleasePhase(rel, apiv1.ReleasePhaseApplying, relLock)
	if err = release.UpdateApplyRelease(storage, rel, params.ExecuteParams.Dryrun, relLock); err != nil {
		return err
	}

	executeOptions = BuildOptions(params.ExecuteParams.Dryrun, m.maxConcurrent)
	var upRel *apiv1.Release
	if upRel, err = engineapi.Apply(ctx, executeOptions, storage, rel, changes, os.Stdout); err != nil {
		return err
	}
	rel = upRel
	// Write resources to DB
	err = m.WriteResources(ctx, rel, stackEntity, specID)
	if err != nil {
		return err
	}
	err = m.ReconcileResources(ctx, stackEntity.ID, rel)
	if err != nil {
		return err
	}

	return nil
}

func (m *StackManager) DestroyStack(ctx context.Context, params *StackRequestParams, w http.ResponseWriter) (err error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Starting destroying stack in StackManager ...")

	// Get the stack entity by id
	stackEntity, err := m.stackRepo.Get(ctx, params.StackID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGettingNonExistingStack
		}
		return err
	}

	// update release to succeeded or failed
	var storage release.Storage
	rel := &apiv1.Release{}
	releaseCreated := false
	defer func() {
		if !releaseCreated {
			return
		}
		if err != nil {
			rel.Phase = apiv1.ReleasePhaseFailed
			_ = release.UpdateDestroyRelease(storage, rel)
			stackEntity.SyncState = constant.StackStateDestroyFailed
		} else {
			rel.Phase = apiv1.ReleasePhaseSucceeded
			err = release.UpdateDestroyRelease(storage, rel)
			// Update LastSyncTimestamp to current time and set stack syncState to synced
			if !params.ExecuteParams.Dryrun {
				stackEntity.SyncState = constant.StackStateDestroySucceeded
			}
		}
		m.stackRepo.Update(ctx, stackEntity)
	}()

	// If the stack is being generated/previewed/applied/destroyed by another request, return an error
	// TODO: This is a temporary solution to prevent multiple requests from destroying the same stack and cause concurrency issues
	// To override this, pass in force == true
	if stackEntity.StackInOperation() && !params.ExecuteParams.Force {
		return ErrStackInOperation
	}

	// Set stack sync state to destroying
	stackEntity.SyncState = constant.StackStateDestroying
	err = m.stackRepo.Update(ctx, stackEntity)
	if err != nil {
		return err
	}

	// create release
	_, stackBackend, project, stack, ws, err := m.metaHelper(ctx, params.StackID, params.Workspace)
	if err != nil {
		return err
	}
	releasePath := getReleasePath(stackEntity.Path, "default")
	storage, err = stackBackend.StateStorageWithPath(releasePath)
	if err != nil {
		return err
	}
	logger.Info("State storage found with path", "releasePath", releasePath)
	if err != nil {
		return err
	}
	rel, err = release.CreateDestroyRelease(storage, project.Name, stack.Name, ws.Name)
	if err != nil {
		return
	}
	if len(rel.Spec.Resources) == 0 {
		return ErrNoManagedResourceToDestroy
	}
	releaseCreated = true

	executeOptions := BuildOptions(params.ExecuteParams.Dryrun, m.maxConcurrent)
	stack.Path = tempPath(stackEntity.Path)

	// compute changes for preview
	changes, err := engineapi.DestroyPreview(executeOptions, rel.Spec, rel.State, project, stack, storage)
	if err != nil {
		return
	}

	// Summary preview table
	changes.Summary(w, true)
	// detail detection
	if params.ExecuteParams.Detail {
		changes.OutputDiff("all")
	}

	// if dryrun, print the hint
	if params.ExecuteParams.Dryrun {
		logger.Info("Dry-run mode enabled, the above resources will be destroyed if dryrun is set to false")
		return ErrDryrunDestroy
	}

	// update release phase to destroying
	rel.Phase = apiv1.ReleasePhaseDestroying
	if err = release.UpdateDestroyRelease(storage, rel); err != nil {
		return
	}
	// Destroy
	logger.Info("Start destroying resources......")
	var upRel *apiv1.Release

	upRel, err = engineapi.Destroy(executeOptions, rel, changes, storage)
	if err != nil {
		return
	}

	// Mark resources as deleted in the database
	err = m.MarkResourcesAsDeleted(ctx, rel)
	if err != nil {
		return err
	}
	rel = upRel
	return nil
}

package release

import (
	"fmt"
	"sync"
	"time"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/log"
)

// GetLatestRelease returns the latest release. If no release exists, return nil.
func GetLatestRelease(storage Storage) (*v1.Release, error) {
	revision := storage.GetLatestRevision()
	if revision == 0 {
		return nil, nil
	}

	r, err := storage.Get(revision)
	if err != nil {
		return nil, err
	}

	return r, err
}

// GetLatestState returns the latest state. If no release exists, return nil.
func GetLatestState(storage Storage) (*v1.State, error) {
	revision := storage.GetLatestRevision()
	if revision == 0 {
		return nil, nil
	}
	r, err := storage.Get(revision)
	if err != nil {
		return nil, err
	}
	return r.State, err
}

// NewApplyRelease news a release object for apply operation, but no creation in the storage.
func NewApplyRelease(storage Storage, project, stack, workspace string) (*v1.Release, error) {
	revision := storage.GetLatestRevision()

	var rel *v1.Release
	currentTime := time.Now()
	if revision == 0 {
		rel = &v1.Release{
			Project:      project,
			Workspace:    workspace,
			Revision:     revision + 1,
			Stack:        stack,
			State:        &v1.State{},
			Phase:        v1.ReleasePhaseGenerating,
			CreateTime:   currentTime,
			ModifiedTime: currentTime,
		}
	} else {
		lastRelease, err := storage.Get(revision)
		if err != nil {
			return nil, err
		}
		if lastRelease.Phase != v1.ReleasePhaseSucceeded && lastRelease.Phase != v1.ReleasePhaseFailed {
			return nil, fmt.Errorf("cannot create a new release of project: %s, workspace: %s. There is a release:%v in progress",
				project, workspace, lastRelease.Revision)
		}

		rel = &v1.Release{
			Project:      project,
			Workspace:    workspace,
			Revision:     revision + 1,
			Stack:        stack,
			State:        lastRelease.State,
			Phase:        v1.ReleasePhaseGenerating,
			CreateTime:   currentTime,
			ModifiedTime: currentTime,
		}
	}

	return rel, nil
}

// NewRollbackRelease news a release object for rollback operation, but no creation in the storage.
func NewRollbackRelease(storage Storage, project, stack, workspace string, revision uint64) (*v1.Release, error) {
	if storage == nil {
		return nil, fmt.Errorf("storage cannot be nil")
	}

	latestRevision := storage.GetLatestRevision()
	if latestRevision <= 0 {
		return nil, fmt.Errorf("invalid latest revision: %d", latestRevision)
	}

	lastRelease, err := storage.Get(latestRevision)
	if err != nil {
		return nil, err
	}
	if lastRelease == nil {
		return nil, fmt.Errorf("last release not found for revision: %d", latestRevision)
	}

	if lastRelease.Phase != v1.ReleasePhaseSucceeded && lastRelease.Phase != v1.ReleasePhaseFailed {
		return nil, fmt.Errorf("cannot create a new release of project: %s, workspace: %s. There is a release:%v in progress",
			project, workspace, lastRelease.Revision)
	}

	if revision <= 0 {
		revision = latestRevision - 1
	}

	rollbackRelease, err := storage.Get(revision)
	if err != nil {
		return nil, err
	}
	if rollbackRelease == nil {
		return nil, fmt.Errorf("rollback release not found for revision: %d", revision)
	}

	if rollbackRelease.Phase != v1.ReleasePhaseSucceeded {
		return nil, fmt.Errorf("cannot create a new rollback release of project: %s, workspace: %s. There is a release:%v not succeeded",
			project, workspace, rollbackRelease.Revision)
	}

	rollbackRelease.Revision = latestRevision + 1
	return rollbackRelease, nil
}

// UpdateApplyRelease updates the release in the storage if dryRun is false. If release phase is failed,
// only logging with no error return.
func UpdateApplyRelease(storage Storage, rel *v1.Release, dryRun bool, relLock *sync.Mutex) error {
	relLock.Lock()
	defer relLock.Unlock()
	if dryRun {
		return nil
	}
	rel.ModifiedTime = time.Now()
	err := storage.Update(rel)
	if rel.Phase == v1.ReleasePhaseFailed && err != nil {
		log.Errorf("failed update release phase to Failed, project %s, workspace %s, revision %d", rel.Project, rel.Workspace, rel.Revision)
		return nil
	}
	return err
}

// CreateDestroyRelease creates a release object in the storage for destroy operation.
func CreateDestroyRelease(storage Storage, project, stack, workspace string) (*v1.Release, error) {
	revision := storage.GetLatestRevision()
	if revision == 0 {
		return nil, fmt.Errorf("cannot find release of project %s, workspace %s", project, workspace)
	}

	lastRelease, err := storage.Get(revision)
	if err != nil {
		return nil, err
	}
	if lastRelease.Phase != v1.ReleasePhaseSucceeded && lastRelease.Phase != v1.ReleasePhaseFailed {
		return nil, fmt.Errorf("cannot create release of project %s, workspace %s cause there is release in progress", project, workspace)
	}

	resources := make([]v1.Resource, len(lastRelease.State.Resources))
	copy(resources, lastRelease.State.Resources)

	secretStore := &v1.SecretStore{}
	if lastRelease.Spec != nil && lastRelease.Spec.SecretStore != nil {
		secretStore = lastRelease.Spec.SecretStore
	}

	specContext := v1.GenericConfig{}
	if lastRelease.Spec != nil && lastRelease.Spec.Context != nil {
		specContext = lastRelease.Spec.Context
	}

	spec := &v1.Spec{Resources: resources, SecretStore: secretStore, Context: specContext}

	// if no resource managed, set phase to Succeeded directly.
	phase := v1.ReleasePhasePreviewing
	if len(resources) == 0 {
		phase = v1.ReleasePhaseSucceeded
	}
	currentTime := time.Now()
	rel := &v1.Release{
		Project:      project,
		Workspace:    workspace,
		Revision:     revision + 1,
		Stack:        stack,
		Spec:         spec,
		State:        lastRelease.State,
		Phase:        phase,
		CreateTime:   currentTime,
		ModifiedTime: currentTime,
	}

	if err = storage.Create(rel); err != nil {
		return nil, fmt.Errorf("create release of project %s workspace %s revision %d failed", project, workspace, rel.Revision)
	}

	return rel, nil
}

// UpdateDestroyRelease updates the release in the storage. If release phase is failed, only logging with
// no error return.
func UpdateDestroyRelease(storage Storage, rel *v1.Release) error {
	rel.ModifiedTime = time.Now()
	err := storage.Update(rel)
	if rel.Phase == v1.ReleasePhaseFailed && err != nil {
		log.Errorf("failed update release phase to Failed, project %s, workspace %s, revision %d", rel.Project, rel.Workspace, rel.Revision)
		return nil
	}
	return err
}

// UpdateReleasePhase updates the release with the specified phase.
func UpdateReleasePhase(rel *v1.Release, phase v1.ReleasePhase, relLock *sync.Mutex) {
	relLock.Lock()
	defer relLock.Unlock()
	rel.Phase = phase
}

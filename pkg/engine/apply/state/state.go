package state

import (
	"sync"
	"time"

	"k8s.io/cli-runtime/pkg/genericiooptions"
	"kusionstack.io/kusion/pkg/engine/release"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/engine/resource/graph"
)

type Metadata struct {
	Project   string
	Stack     string
	Workspace string
}

// State release state
type State struct {
	*Metadata

	// apply state options
	DryRun      bool
	PortForward int
	Watch       bool
	IO          genericiooptions.IOStreams

	// release status data
	CurrentRel *apiv1.Release
	TargetRel  *apiv1.Release
	Gph        *apiv1.Graph
	RelLock    *sync.Mutex

	// port forwarded
	PortForwarded    bool
	PortForwardStop  chan struct{}
	PortForwardReady chan struct{}

	GraphStorage graph.Storage

	// release storage
	ReleaseStorage release.Storage
	ReleaseCreated bool

	// summary
	Ls *LineSummary

	// callback revision
	CallbackRevision uint64
}

func (s *State) CreateStorageRelease(rel *apiv1.Release) error {
	err := s.ReleaseStorage.Create(rel)
	if err != nil {
		return err
	}
	s.ReleaseCreated = true
	return nil
}

func (s *State) UpdateReleasePhaseFailed() (err error) {
	if !s.ReleaseCreated {
		return
	}
	if s.TargetRel == nil || s.ReleaseStorage == nil {
		return nil
	}
	release.UpdateReleasePhase(s.TargetRel, apiv1.ReleasePhaseFailed, s.RelLock)
	if err = release.UpdateApplyRelease(s.ReleaseStorage, s.TargetRel, s.DryRun, s.RelLock); err != nil {
		return
	}
	return nil
}

func (s *State) UpdateReleasePhaseSucceeded() (err error) {
	if !s.ReleaseCreated {
		return
	}
	release.UpdateReleasePhase(s.TargetRel, apiv1.ReleasePhaseSucceeded, s.RelLock)
	if err = release.UpdateApplyRelease(s.ReleaseStorage, s.TargetRel, s.DryRun, s.RelLock); err != nil {
		return
	}
	return nil
}

func (s *State) UpdateReleasePhasePreviewing() (err error) {
	if !s.ReleaseCreated {
		return
	}
	release.UpdateReleasePhase(s.TargetRel, apiv1.ReleasePhasePreviewing, s.RelLock)
	if err = release.UpdateApplyRelease(s.ReleaseStorage, s.TargetRel, s.DryRun, s.RelLock); err != nil {
		return
	}
	return
}

func (s *State) UpdateReleasePhaseApplying() (err error) {
	if !s.ReleaseCreated {
		return
	}
	release.UpdateReleasePhase(s.TargetRel, apiv1.ReleasePhaseApplying, s.RelLock)
	if err = release.UpdateApplyRelease(s.ReleaseStorage, s.TargetRel, s.DryRun, s.RelLock); err != nil {
		return
	}
	return
}

func (s *State) InterruptFunc() {
	if !s.ReleaseCreated {
		return
	}
	release.UpdateReleasePhase(s.TargetRel, apiv1.ReleasePhaseFailed, s.RelLock)
	_ = release.UpdateApplyRelease(s.ReleaseStorage, s.TargetRel, false, s.RelLock)
}

func (s *State) ExitClear() {
	finish := make(chan struct{})

	// clear port forward
	timeOut := time.NewTimer(5 * time.Second)

	go func() {
		if s.PortForwarded {
			s.PortForwardStop <- struct{}{}
		}
		finish <- struct{}{}
	}()
	select {
	case <-timeOut.C:
		return
	case <-finish:
		return
	}
}

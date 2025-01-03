package release

import (
	"fmt"
	"sync"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/engine/resource/graph"
)

type State struct {
	ProjectName   string
	WorkspaceName string

	CurrentRel        *apiv1.Release
	TargetRel         *apiv1.Release
	Gph               *apiv1.Graph
	RelLock           *sync.Mutex
	ReleaseHasStorage bool
	PortForwarded     bool

	GraphStorage   graph.Storage
	ReleaseStorage Storage
}

func NewState(projectName string, workspaceName string) *State {
	return &State{
		ProjectName:    projectName,
		WorkspaceName:  workspaceName,
		RelLock:        &sync.Mutex{},
		GraphStorage:   nil,
		ReleaseStorage: nil,
	}
}

func (s *State) GetReleaseByRevision(revision uint64) (rel *apiv1.Release, err error) {
	if s.ReleaseStorage == nil {
		return nil, fmt.Errorf("GetReleaseByRevision release storage is nil")
	}
	if revision != 0 {
		rel, err = s.ReleaseStorage.Get(revision)
		if err != nil {
			fmt.Printf("No release found for revision %d of project: %s, workspace: %s\n",
				revision, s.ProjectName, s.WorkspaceName)
			return
		}
	} else {
		rel, err = s.ReleaseStorage.Get(s.ReleaseStorage.GetLatestRevision())
		if err != nil {
			fmt.Printf("No release found for project: %s, workspace: %s\n",
				s.ProjectName, s.WorkspaceName)
			return
		}
	}
	return
}

func (s *State) NewReleaseByRevision(revision uint64) (rel *apiv1.Release, err error) {
	if s.ReleaseStorage == nil {
		return nil, fmt.Errorf("NewReleaseByRevision release storage is nil")
	}
	latestRevision := s.ReleaseStorage.GetLatestRevision()

	lastRelease, err := s.ReleaseStorage.Get(latestRevision)
	if err != nil {
		return nil, err
	}
	s.CurrentRel = lastRelease
	if lastRelease.Phase != apiv1.ReleasePhaseSucceeded && lastRelease.Phase != apiv1.ReleasePhaseFailed {
		return nil, fmt.Errorf("cannot create a new release of project: %s, workspace: %s. There is a release:%v in progress",
			s.ProjectName, s.WorkspaceName, lastRelease.Revision)
	}

	rel, err = s.GetReleaseByRevision(revision)
	if err != nil {
		return nil, err
	}
	rel.Revision = lastRelease.Revision + 1
	s.TargetRel = rel
	return
}

func (s *State) CreateStorageRelease(rel *apiv1.Release) error {
	err := s.ReleaseStorage.Create(rel)
	if err != nil {
		s.ReleaseHasStorage = true
	}
	return err
}

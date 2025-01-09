package rel

import (
	"fmt"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/cmd/meta"
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/engine/resource/graph"
	"kusionstack.io/kusion/pkg/workspace"
)

// ... (TestListOptions_Validate remains the same)

func TestListOptions_Run(t *testing.T) {
	opts := &ListOptions{
		MetaOptions: &meta.MetaOptions{
			RefProject: &v1.Project{
				Name: "mock-project",
			},
			RefWorkspace: &v1.Workspace{
				Name: "mock-workspace",
			},
			Backend: &fakeBackendForList{},
		},
	}

	t.Run("No Releases Found", func(t *testing.T) {
		mockey.PatchConvey("mock release storage", t, func() {
			mockStorage := &fakeStorageForList{
				revisions: []uint64{},
				releases:  map[uint64]*v1.Release{},
			}
			mockey.Mock((*fakeBackendForList).ReleaseStorage).
				Return(mockStorage, nil).Build()

			err := opts.Run()
			assert.NoError(t, err)
		})
	})

	// ... (other test cases remain the same)
}

// Fake implementations for testing
type fakeBackendForList struct{}

func (f *fakeBackendForList) ReleaseStorage(project, workspace string) (release.Storage, error) {
	return &fakeStorageForList{}, nil
}

func (f *fakeBackendForList) WorkspaceStorage() (workspace.Storage, error) {
	return &fakeWorkspaceStorage{}, nil
}

func (f *fakeBackendForList) StateStorageWithPath(path string) (release.Storage, error) {
	return &fakeStorageForList{}, nil
}

func (f *fakeBackendForList) GraphStorage(project, workspace string) (graph.Storage, error) {
	return nil, nil
}

func (f *fakeBackendForList) ProjectStorage() (map[string][]string, error) {
	return nil, nil
}

type fakeWorkspaceStorage struct{}

func (f *fakeWorkspaceStorage) Get(name string) (*v1.Workspace, error) {
	return &v1.Workspace{Name: name}, nil
}

func (f *fakeWorkspaceStorage) List() ([]*v1.Workspace, error) {
	return []*v1.Workspace{}, nil
}

func (f *fakeWorkspaceStorage) Create(ws *v1.Workspace) error {
	return nil
}

func (f *fakeWorkspaceStorage) Update(ws *v1.Workspace) error {
	return nil
}

func (f *fakeWorkspaceStorage) Delete(name string) error {
	return nil
}

func (f *fakeWorkspaceStorage) GetCurrent() (string, error) {
	return "current-workspace", nil
}

func (f *fakeWorkspaceStorage) GetNames() ([]string, error) {
	return []string{}, nil
}

func (f *fakeWorkspaceStorage) SetCurrent(name string) error {
	return nil
}

func (f *fakeWorkspaceStorage) RenameWorkspace(oldName, newName string) error {
	return nil
}

type fakeStorageForList struct {
	revisions []uint64
	releases  map[uint64]*v1.Release
}

func (f *fakeStorageForList) Get(revision uint64) (*v1.Release, error) {
	r, ok := f.releases[revision]
	if !ok {
		return nil, fmt.Errorf("release not found")
	}
	return r, nil
}

func (f *fakeStorageForList) GetRevisions() []uint64 {
	return f.revisions
}

func (f *fakeStorageForList) GetLatestRevision() uint64 {
	if len(f.revisions) == 0 {
		return 0
	}
	return f.revisions[len(f.revisions)-1]
}

func (f *fakeStorageForList) Create(release *v1.Release) error {
	return nil
}

func (f *fakeStorageForList) Update(release *v1.Release) error {
	return nil
}

func (f *fakeStorageForList) GetStackBoundRevisions(stack string) []uint64 {
	return f.revisions
}

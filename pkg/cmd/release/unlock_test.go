package rel

import (
	"fmt"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/cmd/meta"
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/project"
	"kusionstack.io/kusion/pkg/workspace"
)

func TestUnlockFlags_ToOptions(t *testing.T) {
	streams := genericiooptions.IOStreams{}

	f := NewUnlockFlags(streams)

	t.Run("Successful Option Creation", func(t *testing.T) {
		mockey.PatchConvey("mock detect project and stack", t, func() {
			mockey.Mock(project.DetectProjectAndStackFrom).Return(&v1.Project{
				Name: "mock-project",
			}, &v1.Stack{
				Name: "mock-stack",
			}, nil).Build()
			_, err := f.ToOptions()
			assert.NoError(t, err)
		})
	})

	t.Run("Failed Option Creation Due to Invalid Backend", func(t *testing.T) {
		s := "invalid-backend"
		f.MetaFlags.Backend = &s
		_, err := f.ToOptions()
		assert.Error(t, err)
	})
}

func TestUnlockOptions_Validate(t *testing.T) {
	opts := &UnlockOptions{}
	streams := genericiooptions.IOStreams{}
	cmd := NewCmdUnlock(streams)

	t.Run("Valid Args", func(t *testing.T) {
		err := opts.Validate(cmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("Invalid Args", func(t *testing.T) {
		err := opts.Validate(cmd, []string{"invalid-args"})
		assert.Error(t, err)
	})
}

func TestUnlockOptions_Run(t *testing.T) {
	opts := &UnlockOptions{
		MetaOptions: &meta.MetaOptions{
			RefProject: &v1.Project{
				Name: "mock-project",
			},
			RefStack: &v1.Stack{
				Name: "mock-stack",
			},
			RefWorkspace: &v1.Workspace{
				Name: "mock-workspace",
			},
			Backend: &fakeBackend{},
		},
	}

	t.Run("Failed to Get Latest Storage Backend", func(t *testing.T) {
		mockey.PatchConvey("mock release storage", t, func() {
			mockey.Mock((*fakeBackend).ReleaseStorage).
				Return(nil, fmt.Errorf("failed to get release storage")).Build()

			err := opts.Run()
			assert.ErrorContains(t, err, "failed to get release storage")
		})
	})

	t.Run("Failed to Get Latest Release", func(t *testing.T) {
		mockey.PatchConvey("mock release storage and release getter", t, func() {
			mockey.Mock((*fakeBackend).ReleaseStorage).
				Return(nil, nil).Build()
			mockey.Mock(release.GetLatestRelease).
				Return(nil, fmt.Errorf("failed to get latest release")).Build()

			err := opts.Run()
			assert.ErrorContains(t, err, "failed to get latest release")
		})
	})

	t.Run("No Release File Found", func(t *testing.T) {
		mockey.PatchConvey("mock release storage and release getter", t, func() {
			mockey.Mock((*fakeBackend).ReleaseStorage).
				Return(nil, nil).Build()
			mockey.Mock(release.GetLatestRelease).
				Return(nil, nil).Build()

			err := opts.Run()
			assert.NoError(t, err)
		})
	})

	t.Run("Failed to Update Release", func(t *testing.T) {
		mockey.PatchConvey("mock release storage and release getter", t, func() {
			mockey.Mock((*fakeBackend).ReleaseStorage).
				Return(&fakeStorage{}, nil).Build()
			mockey.Mock(release.GetLatestRelease).
				Return(&v1.Release{
					Phase: v1.ReleasePhaseApplying,
				}, nil).Build()
			mockey.Mock((*fakeStorage).Update).
				Return(fmt.Errorf("failed to update release")).Build()

			err := opts.Run()
			assert.ErrorContains(t, err, "failed to update release")
		})
	})

	t.Run("Successfully Update Release Phase", func(t *testing.T) {
		mockey.PatchConvey("mock release storage and release getter", t, func() {
			mockey.Mock((*fakeBackend).ReleaseStorage).
				Return(&fakeStorage{}, nil).Build()
			mockey.Mock(release.GetLatestRelease).
				Return(&v1.Release{
					Phase: v1.ReleasePhaseApplying,
				}, nil).Build()

			err := opts.Run()
			assert.NoError(t, err)
		})
	})

	t.Run("No Need to Update Release", func(t *testing.T) {
		mockey.PatchConvey("mock release storage and release getter", t, func() {
			mockey.Mock((*fakeBackend).ReleaseStorage).
				Return(&fakeStorage{}, nil).Build()
			mockey.Mock(release.GetLatestRelease).
				Return(&v1.Release{
					Phase: v1.ReleasePhaseSucceeded,
				}, nil).Build()

			err := opts.Run()
			assert.NoError(t, err)
		})
	})
}

var _ backend.Backend = (*fakeBackend)(nil)

type fakeBackend struct{}

func (f *fakeBackend) WorkspaceStorage() (workspace.Storage, error) {
	return nil, nil
}

func (f *fakeBackend) ReleaseStorage(project, workspace string) (release.Storage, error) {
	return nil, nil
}

var _ release.Storage = (*fakeStorage)(nil)

type fakeStorage struct{}

func (f *fakeStorage) Get(revision uint64) (*v1.Release, error) {
	return nil, nil
}

func (f *fakeStorage) GetRevisions() []uint64 {
	return nil
}

func (f *fakeStorage) GetStackBoundRevisions(stack string) []uint64 {
	return nil
}

func (f *fakeStorage) GetLatestRevision() uint64 {
	return 0
}

func (f *fakeStorage) Create(release *v1.Release) error {
	return nil
}

func (f *fakeStorage) Update(release *v1.Release) error {
	return nil
}

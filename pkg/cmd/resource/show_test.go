package resource

import (
	"fmt"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/engine/resource/graph"
	"kusionstack.io/kusion/pkg/project"
	"kusionstack.io/kusion/pkg/workspace"
)

func TestShowFlags_ToOptions(t *testing.T) {
	streams := genericiooptions.IOStreams{}
	f := NewShowFlags(streams)

	t.Run("Failed Option Creation Due to Missing ID", func(t *testing.T) {
		_, err := f.ToOptions()
		assert.ErrorContains(t, err, "resource ID is required")
	})

	id := "mock-id"
	f.ID = &id

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
		f.Backend = &s
		_, err := f.ToOptions()
		assert.Error(t, err)
	})
}

func TestShowOptions_Validate(t *testing.T) {
	opts := &ShowOptions{}
	streams := genericiooptions.IOStreams{}
	cmd := NewCmdShow(streams)

	t.Run("Valid Args", func(t *testing.T) {
		err := opts.Validate(cmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("Invalid Args", func(t *testing.T) {
		err := opts.Validate(cmd, []string{"invalid-args"})
		assert.Error(t, err)
	})
}

func TestShowOptions_Run(t *testing.T) {
	id := "mock-id"
	projectName := "mock-project"
	workspaceName := "mock-workspace"
	opts := &ShowOptions{
		ID:             &id,
		Project:        &projectName,
		Workspace:      &workspaceName,
		ReleaseStorage: &fakeStorageShow{},
	}

	t.Run("Successfully show the latest release", func(t *testing.T) {
		mockey.PatchConvey("mock release getter", t, func() {
			mockey.Mock((*fakeStorageShow).Get).
				Return(&v1.Release{
					Project:   "mock-project",
					Workspace: "mock-workspace",
					Revision:  1,
					Spec: &v1.Spec{
						Resources: v1.Resources{
							{
								ID:         "mock-id",
								Type:       "",
								Attributes: nil,
								DependsOn:  nil,
								Extensions: nil,
							},
						},
					},
					State: &v1.State{
						Resources: v1.Resources{
							{
								ID:         "mock-id",
								Type:       "",
								Attributes: nil,
								DependsOn:  nil,
								Extensions: nil,
							},
						},
					},
				}, nil).Build()

			err := opts.Run()
			assert.NoError(t, err)
		})
	})

	t.Run("Failed to show the latest release", func(t *testing.T) {
		mockey.PatchConvey("mock release getter", t, func() {
			mockey.Mock((*fakeStorageShow).Get).
				Return(nil, fmt.Errorf("release does not exist")).Build()

			err := opts.Run()
			assert.ErrorContains(t, err, "no release found")
		})
	})

	t.Run("No resources found", func(t *testing.T) {
		mockey.PatchConvey("mock release getter", t, func() {
			mockey.Mock((*fakeStorageShow).Get).
				Return(&v1.Release{
					Project:   "mock-project",
					Workspace: "mock-workspace",
					Revision:  1,
					Spec:      &v1.Spec{},
				}, nil).Build()

			err := opts.Run()
			assert.ErrorContains(t, err, "no resources found")
		})
	})

	t.Run("Resource ID not found", func(t *testing.T) {
		mockey.PatchConvey("mock release getter", t, func() {
			mockey.Mock((*fakeStorageShow).Get).
				Return(&v1.Release{
					Project:   "mock-project",
					Workspace: "mock-workspace",
					Revision:  1,
					Spec: &v1.Spec{
						Resources: v1.Resources{
							{
								ID:         "mock-id-other",
								Type:       "",
								Attributes: nil,
								DependsOn:  nil,
								Extensions: nil,
							},
						},
					},
					State: &v1.State{
						Resources: v1.Resources{
							{
								ID:         "mock-id-other",
								Type:       "",
								Attributes: nil,
								DependsOn:  nil,
								Extensions: nil,
							},
						},
					},
				}, nil).Build()

			err := opts.Run()
			assert.ErrorContains(t, err, "no resource found")
		})
	})
}

var _ backend.Backend = (*fakeBackendShow)(nil)

type fakeBackendShow struct{}

func (f *fakeBackendShow) WorkspaceStorage() (workspace.Storage, error) {
	return nil, nil
}

func (f *fakeBackendShow) ReleaseStorage(_, _ string) (release.Storage, error) {
	return nil, nil
}

func (f *fakeBackendShow) StateStorageWithPath(_ string) (release.Storage, error) {
	return nil, nil
}

func (f *fakeBackendShow) GraphStorage(project, workspace string) (graph.Storage, error) {
	return nil, nil
}

func (f *fakeBackendShow) ProjectStorage() (map[string][]string, error) {
	return nil, nil
}

var _ release.Storage = (*fakeStorageShow)(nil)

type fakeStorageShow struct{}

func (f *fakeStorageShow) Get(_ uint64) (*v1.Release, error) {
	return nil, nil
}

func (f *fakeStorageShow) GetRevisions() []uint64 {
	return nil
}

func (f *fakeStorageShow) GetStackBoundRevisions(_ string) []uint64 {
	return nil
}

func (f *fakeStorageShow) GetLatestRevision() uint64 {
	return 0
}

func (f *fakeStorageShow) Create(_ *v1.Release) error {
	return nil
}

func (f *fakeStorageShow) Update(_ *v1.Release) error {
	return nil
}

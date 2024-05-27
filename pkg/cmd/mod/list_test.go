package mod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/backend/storages"
	"kusionstack.io/kusion/pkg/cmd/meta"
)

func TestFlag_ToOption(t *testing.T) {
	f := &ListFlag{
		MetaFlags: *meta.NewMetaFlags(),
	}

	t.Run("Successful Option Creation", func(t *testing.T) {
		_, err := f.toOption(genericiooptions.IOStreams{})
		assert.NoError(t, err)
	})

	t.Run("Failed Option Creation Due to Invalid Backend", func(t *testing.T) {
		s := "invalid-backend"
		f.MetaFlags.Backend = &s
		_, err := f.toOption(genericiooptions.IOStreams{})
		assert.Error(t, err)
	})
}

func TestListOptions_Run(t *testing.T) {
	o := &ListOption{
		Workspace: &v1.Workspace{
			Modules: map[string]*v1.ModuleConfig{
				"module1": {
					Path:    "path1",
					Version: "v1.0.0",
				},
			},
		},
		StorageBackend: &storages.LocalStorage{},
		IO:             genericiooptions.IOStreams{},
	}

	t.Run("Successful Run", func(t *testing.T) {
		err := o.Run()
		assert.NoError(t, err)
	})

	t.Run("Failed Run Due to Nil Workspace", func(t *testing.T) {
		o.Workspace = nil
		err := o.Run()
		assert.Error(t, err)
	})
}

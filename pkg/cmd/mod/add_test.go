package mod

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"kcl-lang.io/kpm/pkg/api"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/cmd/meta"
)

func TestAddOptions_Run(t *testing.T) {
	t.Run("EmptyWorkspaceName", func(t *testing.T) {
		o := &AddOptions{
			MetaOptions: meta.MetaOptions{},
			ModuleName:  "testModule",
			IO:          genericiooptions.IOStreams{},
		}
		err := o.Run()
		assert.Error(t, err)
	})

	t.Run("ModuleNotFoundInWorkspace", func(t *testing.T) {
		o := &AddOptions{
			MetaOptions: meta.MetaOptions{
				RefWorkspace: &v1.Workspace{
					Name: "testWorkspace",
				},
			},
			ModuleName: "testModule",
			IO:         genericiooptions.IOStreams{},
		}
		err := o.Run()
		assert.Error(t, err)
	})

	t.Run("InvalidModulePath", func(t *testing.T) {
		o := &AddOptions{
			MetaOptions: meta.MetaOptions{
				RefStack: &v1.Stack{
					Path: "./testdata/dev",
				},
				RefWorkspace: &v1.Workspace{
					Name: "testWorkspace",
					Modules: v1.ModuleConfigs{
						"testModule": {
							Path: "invalidPath",
						},
					},
				},
			},
			ModuleName: "testModule",
			IO:         genericiooptions.IOStreams{},
		}
		err := o.Run()
		assert.Error(t, err)
	})

	t.Run("ValidModuleAddition", func(t *testing.T) {
		b := make([]byte, 4)
		letter := "abcdefghijklmnopqrstuvwxyz"
		for i := range b {
			b[i] = letter[rand.Intn(len(letter))]
		}

		o := &AddOptions{
			MetaOptions: meta.MetaOptions{
				RefStack: &v1.Stack{
					Path: "./testdata/dev",
				},
				RefWorkspace: &v1.Workspace{
					Name: "dev",
					Modules: v1.ModuleConfigs{
						"service": {
							Path:    "oci://ghcr.io/kusionstack/service",
							Version: string(b),
						},
					},
				},
			},
			ModuleName: "service",
			IO:         genericiooptions.IOStreams{},
		}
		err := o.Run()
		assert.NoError(t, err)
		kclPkg, err := api.GetKclPackage(o.RefStack.Path)
		assert.NoError(t, err)
		file := kclPkg.GetDependenciesInModFile()
		assert.Equal(t, file.Deps["service"].Version, string(b))
	})
}

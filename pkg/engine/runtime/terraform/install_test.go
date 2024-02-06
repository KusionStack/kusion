package terraform

import (
	"fmt"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

func TestCLIInstaller_CheckAndInstall(t *testing.T) {
	mockey.PatchConvey("NoResources", t, func() {
		installer := &CLIInstaller{
			Intent: &v1.Intent{
				Resources: v1.Resources{},
			},
		}
		err := installer.CheckAndInstall()
		assert.Nil(t, err)
	})

	mockey.PatchConvey("NoTerraformResources", t, func() {
		installer := &CLIInstaller{
			Intent: &v1.Intent{
				Resources: v1.Resources{
					v1.Resource{
						Type: v1.Kubernetes,
					},
				},
			},
		}
		err := installer.CheckAndInstall()
		assert.Nil(t, err)
	})

	mockey.PatchConvey("ExistingTerraformExecutable", t, func() {
		mockey.Mock(checkTerraformExecutable).To(func() error {
			return nil
		}).Build()
		installer := &CLIInstaller{
			Intent: &v1.Intent{
				Resources: v1.Resources{
					v1.Resource{
						Type: v1.Terraform,
					},
				},
			},
		}
		err := installer.CheckAndInstall()
		assert.Nil(t, err)
	})

	mockey.PatchConvey("InstallTerraformTimeout", t, func() {
		mockey.Mock(checkTerraformExecutable).To(func() error {
			return fmt.Errorf("terraform executable not found")
		}).Build()
		mockey.Mock(installTerraform).To(func() error {
			return fmt.Errorf("install timeout")
		}).Build()
		installer := &CLIInstaller{
			Intent: &v1.Intent{
				Resources: v1.Resources{
					v1.Resource{
						Type: v1.Terraform,
					},
				},
			},
		}
		err := installer.CheckAndInstall()
		assert.ErrorContains(t, err, "install timeout")
	})

	mockey.PatchConvey("SuccessfullyInstalled", t, func() {
		mockey.Mock(checkTerraformExecutable).To(func() error {
			return fmt.Errorf("terraform executable not found")
		}).Build()
		mockey.Mock(installTerraform).To(func() error {
			return nil
		}).Build()
		installer := &CLIInstaller{
			Intent: &v1.Intent{
				Resources: v1.Resources{
					v1.Resource{
						Type: v1.Terraform,
					},
				},
			},
		}
		err := installer.CheckAndInstall()
		assert.Nil(t, err)
	})
}

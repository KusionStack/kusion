package terraform

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util/kfile"
)

var (
	tfInstallSubDir  = "terraform"
	tfInstallTimeout = 3 * time.Minute
)

type CLIInstaller struct {
	Intent *apiv1.Intent
}

// Check and install the terraform executable binary if it has not been downloaded.
func (installer *CLIInstaller) CheckAndInstall() error {
	if len(installer.Intent.Resources) < 1 {
		return nil
	}

	for i, res := range installer.Intent.Resources {
		if res.Type == apiv1.Terraform {
			break
		}

		if i == installer.Intent.Resources.Len()-1 {
			return nil
		}
	}

	if err := checkTerraformExecutable(); err != nil {
		log.Warn("Terraform executable binary is not found")

		if err := installTerraform(); err != nil {
			return err
		}

		log.Info("Successfully installed terraform and set the executable path")
	}

	return nil
}

// check whether the terraform executable binary has been installed.
func checkTerraformExecutable() error {
	// select the executable file name according to the operating system.
	var executable string
	if runtime.GOOS == "windows" {
		executable = "terraform.exe"
	} else {
		executable = "terraform"
	}

	if err := exec.Command(executable, "--version").Run(); err == nil {
		return nil
	}

	installDir, err := getTerraformInstallDir()
	if err != nil {
		return err
	}

	execPath := filepath.Join(installDir, executable)
	if err := exec.Command(execPath, "--version").Run(); err != nil {
		return err
	}

	return setTerraformExecPathEnv(execPath)
}

// install and set the environment variable of executable path for terraform binary,
// currently the latest version will be downloaded by default.
func installTerraform() error {
	log.Info("Installing terraform binary with the latest version...")

	installDir, err := getTerraformInstallDir()
	if err != nil {
		return err
	}

	installer := &releases.LatestVersion{
		Product:    product.Terraform,
		InstallDir: installDir,
		Timeout:    tfInstallTimeout,
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		return err
	}

	log.Infof("Successfully located terraform binary: %s\n", execPath)

	return setTerraformExecPathEnv(execPath)
}

// set the environment variable of executable path for terraform binary,
// note that this env only takes effect for the current process.
func setTerraformExecPathEnv(execPath string) error {
	if execPath == "" {
		return fmt.Errorf("empty executable path for terraform binary")
	}

	log.Info("Setting the environment variable of executable path for terraform...")
	currentPath := os.Getenv("PATH")

	// select the path separator according to the operating system.
	var pathSeparator string
	if runtime.GOOS == "windows" {
		pathSeparator = ";"
	} else {
		pathSeparator = ":"
	}

	newPath := filepath.Dir(execPath) + pathSeparator + currentPath

	return os.Setenv("PATH", newPath)
}

// get the installation directory for terraform binary, and by default
// it is ~/.kusion/terraform.
func getTerraformInstallDir() (string, error) {
	kusionDir, err := kfile.KusionDataFolder()
	if err != nil {
		return "", err
	}

	installDir := filepath.Join(kusionDir, tfInstallSubDir)

	if _, err = os.Stat(installDir); os.IsNotExist(err) {
		if err := os.Mkdir(installDir, 0o755); err != nil {
			return "", fmt.Errorf("failed to create terraform install directory: %v", err)
		}
	}

	return installDir, nil
}

package compile

import (
	"os"
	"os/exec"
	"path/filepath"

	"kusionstack.io/kusion/pkg/util/io"
)

const KUSION_KCL_PATH_ENV = "KUSION_KCL_PATH"

var kclAppPath = getKclPath()

func getKclPath() string {
	// 1. try ${KUSION_KCL_PATH_ENV}

	if kclPath := os.Getenv(KUSION_KCL_PATH_ENV); kclPath != "" {
		return kclPath
	}

	// 2.1 try ${appPath}/kclvm/bin/kcl
	// 2.2 try ${appPath}/../kclvm/bin/kcl
	// 2.3 try ${PWD}/kclvm/bin/kcl

	var kclPathList []string
	if appPath, _ := os.Executable(); appPath != "" {
		kclPathList = append(kclPathList,
			filepath.Join(filepath.Dir(appPath), "kclvm", "bin", "kcl"),
			filepath.Join(filepath.Dir(filepath.Dir(appPath)), "kclvm", "bin", "kcl"),
		)
	}
	if workDir, _ := os.Getwd(); workDir != "" {
		kclPathList = append(kclPathList,
			filepath.Join(workDir, "kclvm", "bin", "kcl"),
		)
	}
	for _, kclPath := range kclPathList {
		if ok, _ := io.IsFile(kclPath); ok {
			return kclPath
		}
	}

	// 3. try ${PATH}/kcl

	if kclPath, _ := exec.LookPath("kcl"); kclPath != "" {
		return kclPath
	}

	return "kcl"
}

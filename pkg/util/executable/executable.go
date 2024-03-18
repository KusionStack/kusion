package executable

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const unableToFindExecutableProgram = "unable to find executable program: %s"

// FindExecutable attempts to find the needed executable in various locations on the
// filesystem, eventually resorting to searching in $PATH.
func FindExecutable(program string) (string, error) {
	if runtime.GOOS == "windows" && !strings.HasSuffix(program, ".exe") {
		program += ".exe"
	}
	// look in the same directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("unable to get current working directory: %w", err)
	}

	cwdProgram := filepath.Join(cwd, program)
	if fileInfo, err := os.Stat(cwdProgram); !os.IsNotExist(err) && !fileInfo.Mode().IsDir() {
		return cwdProgram, nil
	}

	// look in potentials $GOPATH/bin
	if goPath := os.Getenv("GOPATH"); len(goPath) > 0 {
		// splitGoPath will return a list of paths in which to look for the binary.
		// Because the GOPATH can take the form of multiple paths (https://golang.org/cmd/go/#hdr-GOPATH_environment_variable)
		// we need to split the GOPATH, and look into each of the paths.
		// If the GOPATH hold only one path, there will only be one element in the slice.
		pathParts := splitGoPath(goPath, runtime.GOOS)
		for _, pp := range pathParts {
			goProgramPath := filepath.Join(pp, "bin", program)
			fileInfo, err := os.Stat(goProgramPath)
			if err != nil && !os.IsNotExist(err) {
				return "", err
			}

			if fileInfo != nil && !fileInfo.Mode().IsDir() {
				return goProgramPath, nil
			}
		}
	}

	// look in the $PATH somewhere
	if fullPath, err := exec.LookPath(program); err == nil {
		return fullPath, nil
	}

	return "", fmt.Errorf(unableToFindExecutableProgram, program)
}

func splitGoPath(goPath string, os string) []string {
	var sep string
	switch os {
	case "windows":
		sep = ";"
	case "linux", "darwin":
		sep = ":"
	}

	return strings.Split(goPath, sep)
}

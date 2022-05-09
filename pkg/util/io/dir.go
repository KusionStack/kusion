package io

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

// Create dir recursively if not exist
// Note: input argument must be a file path
func CreateDirIfNotExist(filePath string) error {
	fileDir := path.Dir(filePath)
	_, err := os.Stat(fileDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(fileDir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

// OutDir returns an absolute representation of path after dir check
// Returns absolute path including trailing '/' or error if path does not exist.
func OutDir(path string) (string, error) {
	outDir, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	stat, err := os.Stat(outDir)
	if err != nil {
		return "", err
	}

	if !stat.IsDir() {
		return "", fmt.Errorf("output directory %s is not a directory", outDir)
	}
	outDir += "/"
	return outDir, nil
}

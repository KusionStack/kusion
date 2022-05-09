package io

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
)

const (
	Slash = "/"
)

// IsFileOrDirExist checks whether a file or a dir exists
func IsFileOrDirExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && !os.IsExist(err) {
		return false
	}
	return true
}

// IsDir checks whether the path is a dir
func IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil && !os.IsExist(err) {
		return false, fmt.Errorf("file or dir with path %s doesn't exist", path)
	}
	return info.IsDir(), nil
}

// IsFile checks whether the path is a file
func IsFile(path string) (bool, error) {
	ok, err := IsDir(path)
	if err != nil {
		return false, fmt.Errorf("file or dir with path %s doesn't exist", path)
	}
	return !ok, nil
}

// RenamePath renames (moves) oldPath to newPath, and creates needed directories in newPath
// If newPath already exists, RenamePath will return an error
func RenamePath(oldPath, newPath string) error {
	if !IsFileOrDirExist(oldPath) {
		return fmt.Errorf("oldpath %s doesn't exist", oldPath)
	}
	if IsFileOrDirExist(newPath) {
		return fmt.Errorf("newpath %s already exists", newPath)
	}

	newPathWithoutSlash := strings.TrimSuffix(newPath, Slash)
	lastSlashIndex := strings.LastIndex(newPathWithoutSlash, Slash)
	if lastSlashIndex == -1 {
		return fmt.Errorf("format of newpath %s is wrong", newPath)
	}
	newDir := newPath[:lastSlashIndex]
	if !IsFileOrDirExist(newDir) {
		err := os.MkdirAll(newDir, os.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "make directory %s failed", newDir)
		}
	}

	err := os.Rename(oldPath, newPath)
	if err != nil {
		return errors.Wrapf(err, "rename oldpath %s to newpath %s failed", oldPath, newPath)
	}

	return nil
}

// IsBinary returns true if a zero byte occurs within the first
// 8000 bytes (or the entire length if shorter). This is the
// same approach that git uses to determine if a file is binary.
func IsBinary(bytes []byte) bool {
	const firstFewBytes = 8000

	length := len(bytes)
	if firstFewBytes < length {
		length = firstFewBytes
	}

	for i := 0; i < length; i++ {
		if bytes[i] == 0 {
			return true
		}
	}

	return false
}

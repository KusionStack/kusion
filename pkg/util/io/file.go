package io

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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

// CopyFile copies the file at source to dest
func CopyFile(source, dest string) error {
	sf, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("unable to open source file [%s]: %q", source, err)
	}
	defer sf.Close()
	fi, err := sf.Stat()
	if err != nil {
		return fmt.Errorf("unable to stat source file [%s]: %q", source, err)
	}

	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("unable to create directory [%s]: %q", dir, err)
	}
	df, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("unable to create destination file [%s]: %q", dest, err)
	}
	defer df.Close()

	_, err = io.Copy(df, sf)
	if err != nil {
		return fmt.Errorf("unable to copy [%s] to [%s]: %q", source, dest, err)
	}

	if err := os.Chmod(dest, fi.Mode()); err != nil {
		return fmt.Errorf("unable to close destination file: %q", err)
	}
	return nil
}

// SameFile returns true if the two given paths refer to the same physical
// file on disk, using the unique file identifiers from the underlying
// operating system. For example, on Unix systems this checks whether the
// two files are on the same device and have the same inode.
func SameFile(a, b string) (bool, error) {
	if a == b {
		return true, nil
	}

	aInfo, err := os.Lstat(a)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	bInfo, err := os.Lstat(b)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return os.SameFile(aInfo, bInfo), nil
}

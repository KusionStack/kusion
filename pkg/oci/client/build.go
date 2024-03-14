package client

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fluxcd/pkg/sourceignore"
)

// Build archives the given directory as a tarball to the given local path.
// While archiving, any environment specific data (for example, the user and group name) is stripped from file headers.
// Original copy from https://github.com/fluxcd/pkg/blob/main/oci/client/build.go
func (c *Client) Build(artifactPath, sourceDir string, ignorePaths []string) (err error) {
	return build(artifactPath, sourceDir, ignorePaths)
}

func build(artifactPath, sourceDir string, ignorePaths []string) (err error) {
	absDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return err
	}

	dirStat, err := os.Stat(absDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("invalid source dir path: %s", absDir)
	}

	tf, err := os.CreateTemp(filepath.Split(absDir))
	if err != nil {
		return err
	}
	tmpName := tf.Name()
	defer func() {
		if err != nil {
			os.Remove(tmpName)
		}
	}()

	ignore := strings.Join(ignorePaths, "\n")
	domain := strings.Split(filepath.Clean(absDir), string(filepath.Separator))
	ignorePatterns := sourceignore.ReadPatterns(strings.NewReader(ignore), domain)
	matcher := sourceignore.NewMatcher(ignorePatterns)
	filter := func(p string, fi os.FileInfo) bool {
		return matcher.Match(strings.Split(p, string(filepath.Separator)), fi.IsDir())
	}

	sz := &writeCounter{}
	mw := io.MultiWriter(tf, sz)

	gw := gzip.NewWriter(mw)
	tw := tar.NewWriter(gw)
	if err := filepath.Walk(absDir, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Ignore anything that is not a file or directories e.g. symlinks
		if m := fi.Mode(); !(m.IsRegular() || m.IsDir()) {
			return nil
		}

		if len(ignorePatterns) > 0 && filter(p, fi) {
			return nil
		}

		header, err := tar.FileInfoHeader(fi, p)
		if err != nil {
			return err
		}
		if dirStat.IsDir() {
			// The name needs to be modified to maintain directory structure
			// as tar.FileInfoHeader only has access to the base name of the file.
			// Ref: https://golang.org/src/archive/tar/common.go?#L6264
			//
			// we only want to do this if a directory was passed in
			relFilePath, err := filepath.Rel(absDir, p)
			if err != nil {
				return err
			}
			// Normalize file path so it works on windows
			header.Name = filepath.ToSlash(relFilePath)
		}

		// Remove any environment specific data.
		header.Gid = 0
		header.Uid = 0
		header.Uname = ""
		header.Gname = ""
		header.ModTime = time.Time{}
		header.AccessTime = time.Time{}
		header.ChangeTime = time.Time{}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}
		f, err := os.Open(p)
		if err != nil {
			f.Close()
			return err
		}
		if _, err := io.Copy(tw, f); err != nil {
			f.Close()
			return err
		}
		return f.Close()
	}); err != nil {
		tw.Close()
		gw.Close()
		tf.Close()
		return err
	}

	if err := tw.Close(); err != nil {
		gw.Close()
		tf.Close()
		return err
	}
	if err := gw.Close(); err != nil {
		tf.Close()
		return err
	}
	if err := tf.Close(); err != nil {
		return err
	}

	if err := os.Chmod(tmpName, 0o640); err != nil {
		return err
	}

	return renameWithFallback(tmpName, artifactPath)
}

func renameWithFallback(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	f2, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f2.Close()

	_, err = io.Copy(f2, f)
	if err != nil {
		return err
	}
	return nil
}

type writeCounter struct {
	written int64
}

func (wc *writeCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.written += int64(n)
	return n, nil
}

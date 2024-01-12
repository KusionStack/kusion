package scaffold

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

var (
	//go:embed internal
	internalTemplates embed.FS

	internalDir = "internal"
)

// GenInternalTemplates save localTemplates(FS) to internal-templates(target directory).
func GenInternalTemplates() error {
	baseTemplateDir, err := GetTemplateDir(BaseTemplateDir)
	if err != nil {
		return err
	}
	return fs.WalkDir(internalTemplates, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			destDir := filepath.Join(baseTemplateDir, path)
			err := os.MkdirAll(destDir, os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			bytes, err := internalTemplates.ReadFile(path)
			if err != nil {
				return err
			}
			destFile := filepath.Join(baseTemplateDir, path)
			err = writeAllBytes(destFile, bytes, true, os.ModePerm)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// GetInternalTemplates exports internal templates which is embed in binary.
func GetInternalTemplates() embed.FS {
	return internalTemplates
}

// Transfer embed.FS into afero.Fs.
func Transfer(srcFS embed.FS) (afero.Fs, error) {
	destFS := afero.NewMemMapFs()
	return destFS, fs.WalkDir(srcFS, internalDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if err = destFS.MkdirAll(path, DefaultDirectoryPermission); err != nil {
				return err
			}
		} else {
			// Read source file content
			content, err := fs.ReadFile(internalTemplates, path)
			if err != nil {
				return err
			}
			// Create or Update
			writer, err := destFS.OpenFile(path, CreateOrUpdate, DefaultFilePermission)
			if err != nil {
				return err
			}
			defer func() {
				if closeErr := writer.Close(); err == nil && closeErr != nil {
					err = closeErr
				}
			}()
			// Write into FS
			if _, err := writer.Write(content); err != nil {
				return err
			}
		}
		return nil
	})
}

// InternalTemplateNameToPath return a map of template name to path.
func InternalTemplateNameToPath() map[string]string {
	schemaToPath := make(map[string]string)
	dirs, err := fs.ReadDir(internalTemplates, internalDir)
	if err != nil {
		return schemaToPath
	}
	for _, dir := range dirs {
		if dir.IsDir() {
			schemaToPath[dir.Name()] = filepath.Join(internalDir, dir.Name())
		}
	}
	return schemaToPath
}

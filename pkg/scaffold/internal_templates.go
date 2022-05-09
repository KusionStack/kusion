package scaffold

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed internal
var localTemplates embed.FS

// GenInternalTemplates save localTemplates(FS) to internal-templates(target directory)
func GenInternalTemplates() error {
	baseTemplateDir, err := GetTemplateDir(BaseTemplateDir)
	if err != nil {
		return err
	}
	return fs.WalkDir(localTemplates, ".", func(path string, d fs.DirEntry, err error) error {
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
			bytes, err := localTemplates.ReadFile(path)
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

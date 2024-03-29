package scaffold

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"

	"kusionstack.io/kusion/pkg/backend"
)

const demoTmplDir = "quickstart"

//go:embed quickstart
var demoFS embed.FS

// GenDemoProject creates the demo project with a specified name in the specified directory.
func GenDemoProject(dir, name string) error {
	// Init default workspace for the initialized demo project if not exists.
	_, err := backend.NewWorkspaceStorage("")
	if err != nil {
		return err
	}

	// Define the embeded template parameter data.
	data := struct {
		ProjectName string
	}{
		ProjectName: name,
	}

	// Walk through the embeded template and creates the demo project with the specified name in the specified directory.
	err = fs.WalkDir(demoFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the top-level root directory of the embeded template.
		relPath, err := filepath.Rel(demoTmplDir, path)
		if err != nil {
			return err
		}
		if relPath == "" || relPath == "." {
			return nil
		}

		dstPath := filepath.Join(dir, relPath)
		if d.IsDir() {
			if err := os.MkdirAll(dstPath, os.ModePerm); err != nil {
				return err
			}
		} else {
			srcFile, err := demoFS.ReadFile(path)
			if err != nil {
				return err
			}

			dstFile, err := os.Create(dstPath)
			if err != nil {
				return err
			}
			defer dstFile.Close()

			tmpl, err := template.New(filepath.Base(path)).Parse(string(srcFile))
			if err != nil {
				return err
			}

			if err = tmpl.Execute(dstFile, data); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

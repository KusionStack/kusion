package scaffold

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestGenInternalTemplates(t *testing.T) {
	targetDir, err := GetTemplateDir(InternalTemplateDir)
	assert.Nil(t, err)
	err = GenInternalTemplates()
	assert.Nil(t, err)
	srcDir := "internal"

	var checkFiles func(src, target string) error
	checkFiles = func(src, target string) error {
		srcFileInfos, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		targetFileInfos, err := os.ReadDir(target)
		if err != nil {
			return err
		}
		for i := range srcFileInfos {
			srcFileInfo := srcFileInfos[i]
			if srcFileInfo.Name() == KusionYaml {
				// kusion.yaml is not rendered
				continue
			}
			targetFileInfo := targetFileInfos[i]
			assert.Equal(t, srcFileInfo.IsDir(), targetFileInfo.IsDir())
			assert.Equal(t, srcFileInfo.Name(), targetFileInfo.Name())
			srcPath := filepath.Join(src, srcFileInfo.Name())
			targetPath := filepath.Join(target, targetFileInfo.Name())
			if srcFileInfo.IsDir() && targetFileInfo.IsDir() {
				// recursive check
				return checkFiles(srcPath, targetPath)
			} else if !srcFileInfo.IsDir() && !targetFileInfo.IsDir() {
				// read content
				srcBytes, err := os.ReadFile(srcPath)
				if err != nil {
					return err
				}
				targetBytes, err := os.ReadFile(targetPath)
				if err != nil {
					return err
				}
				assert.Equal(t, srcBytes, targetBytes)
			}
		}
		return nil
	}
	// check files tree
	err = checkFiles(srcDir, targetDir)
	assert.Nil(t, err)
}

func TestInternalSchemas(t *testing.T) {
	schemas := InternalTemplateNameToPath()
	path, ok := schemas[templateName]
	assert.True(t, ok)
	assert.Equal(t, path, filepath.Join(templateDir, templateName))
}

func Test_readIntoFS(t *testing.T) {
	local := afero.NewMemMapFs()
	err := ReadTemplate(internalDir, local)
	assert.Nil(t, err)

	got := []string{}
	err = afero.Walk(local, filepath.Join(templateDir, templateName), func(path string, info fs.FileInfo, err error) error {
		got = append(got, path)
		return nil
	})
	assert.Nil(t, err)

	want := []string{
		"internal/single-stack-sample",
		"internal/single-stack-sample/README.md",
		"internal/single-stack-sample/dev",
		"internal/single-stack-sample/dev/kcl.mod",
		"internal/single-stack-sample/dev/kcl.mod.lock",
		"internal/single-stack-sample/dev/main.k",
		"internal/single-stack-sample/dev/stack.yaml",
		"internal/single-stack-sample/kusion.yaml",
		"internal/single-stack-sample/project.yaml",
	}
	assert.Equal(t, want, got)
}

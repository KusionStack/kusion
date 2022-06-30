package scaffold

import (
	"io/fs"
	"io/ioutil"
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
		srcFileInfos, err := ioutil.ReadDir(src)
		if err != nil {
			return err
		}
		targetFileInfos, err := ioutil.ReadDir(target)
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
				srcBytes, err := ioutil.ReadFile(srcPath)
				if err != nil {
					return err
				}
				targetBytes, err := ioutil.ReadFile(targetPath)
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
		"internal/deployment-single-stack",
		"internal/deployment-single-stack/README.md",
		"internal/deployment-single-stack/base",
		"internal/deployment-single-stack/base/base.k",
		"internal/deployment-single-stack/dev",
		"internal/deployment-single-stack/dev/ci-test",
		"internal/deployment-single-stack/dev/ci-test/settings.yaml",
		"internal/deployment-single-stack/dev/kcl.yaml",
		"internal/deployment-single-stack/dev/main.k",
		"internal/deployment-single-stack/dev/stack.yaml",
		"internal/deployment-single-stack/kusion.yaml",
		"internal/deployment-single-stack/project.yaml",
	}
	assert.Equal(t, want, got)
}

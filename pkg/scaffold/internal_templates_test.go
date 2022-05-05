package scaffold

import (
	"io/ioutil"
	"path/filepath"
	"testing"

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

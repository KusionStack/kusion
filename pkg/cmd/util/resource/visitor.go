package resource

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/yaml"
)

var FileExtensions = []string{".yaml", ".yml"}

type CrdVisitor struct {
	Path string
}

// Visit walks a list of resources under target path.
func (v *CrdVisitor) Visit() (objs []map[string]interface{}, err error) {
	err = filepath.WalkDir(v.Path, func(filePath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// check file extension
		if ignoreFile(filePath, FileExtensions) {
			return nil
		}

		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		decoder := yaml.NewYAMLOrJSONDecoder(f, 4096)
		for {
			data := make(map[string]interface{})
			if err := decoder.Decode(&data); err != nil {
				if err == io.EOF {
					return nil
				}
				return fmt.Errorf("error parsing %s: %v", filePath, err)
			}
			if len(data) == 0 {
				continue
			}

			objs = append(objs, data)
		}
	})
	return objs, err
}

func ignoreFile(path string, extensions []string) bool {
	if len(extensions) == 0 {
		return false
	}
	ext := filepath.Ext(path)
	for _, s := range extensions {
		if s == ext {
			return false
		}
	}
	return true
}

package crd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/util/yaml"

	"kusionstack.io/kusion/pkg/resources"
)

const (
	Directory = "crd"
)

var FileExtensions = []string{".yaml", ".yml", ".json"}

type crdVisitor struct {
	Path string
}

func NewVisitor(path string) resources.Visitor {
	return &crdVisitor{Path: path}
}

// Visit read all YAML files under target path
func (v *crdVisitor) Visit() (objs []interface{}, err error) {
	err = filepath.WalkDir(v.Path, func(filePath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// check file extension
		// todo dayuan validate yaml content to make sure it is a k8s CRD resource

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

// ignoreFile indicates a filename is ended with specified extension or not
func ignoreFile(path string, extensions []string) bool {
	if len(extensions) == 0 {
		return false
	}
	ext := filepath.Ext(path)
	for _, s := range extensions {
		if strings.EqualFold(s, ext) {
			return false
		}
	}
	return true
}

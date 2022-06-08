package scaffold

import (
	"errors"
	"io/ioutil"
	"sync"

	"gopkg.in/yaml.v3"
)

// projectTemplateSingleton is a singleton instance of projectTemplateLoader, which controls a global map of instances of ProjectTemplate
// configs (one per path).
var projectTemplateSingleton = &projectTemplateLoader{
	internal: map[string]*ProjectTemplate{},
}

// projectTemplateLoader is used to load a single global instance of a ProjectTemplate config.
type projectTemplateLoader struct {
	sync.RWMutex
	internal map[string]*ProjectTemplate
}

// LoadProjectTemplate reads a project definition from a file.
func LoadProjectTemplate(path string) (*ProjectTemplate, error) {
	if path == "" {
		return nil, errors.New("path is empty")
	}

	return projectTemplateSingleton.load(path)
}

// Load a ProjectTemplate config file from the specified path. The configuration will be cached for subsequent loads.
func (singleton *projectTemplateLoader) load(path string) (*ProjectTemplate, error) {
	singleton.Lock()
	defer singleton.Unlock()

	if v, ok := singleton.internal[path]; ok {
		return v, nil
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var project ProjectTemplate
	err = yaml.Unmarshal(b, &project)
	if err != nil {
		return nil, err
	}

	singleton.internal[path] = &project
	return &project, nil
}

package project

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/fsutil"
	yamlv3 "gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/util/sets"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/log"
)

var (
	ErrNotProjectDirectory = errors.New("path must be a project directory")
	ErrProjectNotUnique    = errors.New("the project obtained is not unique")
	ErrNotStackDirectory   = errors.New("path must be a stack directory")
	ErrStackNotUnique      = errors.New("the stack obtained is not unique")
)

const (
	ProjectFile = "project.yaml"
	StackFile   = "stack.yaml"
)

// DetectProjectAndStack try to get stack and project from given path
func DetectProjectAndStack(stackDir string) (p *v1.Project, s *v1.Stack, err error) {
	stackDir, err = filepath.Abs(stackDir)
	if err != nil {
		return nil, nil, err
	}

	s, err = GetStackFrom(stackDir)
	if err != nil {
		return nil, nil, err
	}

	projectDir, err := findProjectPathFrom(stackDir)
	if err != nil {
		return nil, nil, err
	}

	p, err = getProjectFrom(projectDir)
	if err != nil {
		return nil, nil, err
	}

	return p, s, nil
}

// isProjectFile determine whether the given path is Project file
func isProjectFile(path string) bool {
	f, err := os.Stat(path)
	return err == nil && !f.IsDir() && f.Mode().IsRegular() && filepath.Base(path) == ProjectFile
}

// isProject determine whether the given path is Project directory
func isProject(path string) bool {
	f, err := os.Stat(path)
	f2, err2 := os.Stat(filepath.Join(path, ProjectFile))

	if (err == nil && f.IsDir()) && (err2 == nil && f2.Mode().IsRegular()) {
		return true
	}

	return false
}

// getProjectFrom get project from the given path
func getProjectFrom(path string) (*v1.Project, error) {
	if !isProject(path) {
		return nil, ErrNotProjectDirectory
	}

	projects, err := FindAllProjectsFrom(path)
	if err != nil {
		return nil, err
	}

	if len(projects) != 1 {
		return nil, ErrProjectNotUnique
	}

	return projects[0], nil
}

// findProjectPathFrom locates the closest project from the given path, searching "upwards" in the directory
// hierarchy. If no project is found, an empty path is returned.
func findProjectPathFrom(path string) (string, error) {
	file, err := fsutil.WalkUp(path, isProjectFile, func(s string) bool {
		return true
	})
	if err != nil {
		return "", err
	}

	return filepath.Dir(file), nil
}

// FindAllProjectsFrom find all project from the given path
func FindAllProjectsFrom(path string) ([]*v1.Project, error) {
	var projects []*v1.Project
	s := sets.NewString()
	err := filepath.WalkDir(path, func(p string, _ fs.DirEntry, _ error) error {
		if isProject(p) && !s.Has(p) {
			// Parse project.yaml
			project, err := parseProjectYamlFile(p)
			if err != nil {
				log.Error(err)
				return fmt.Errorf("parse project.yaml failed. %w", err)
			}

			// Find all stacks
			stacks, err := FindAllStacksFrom(p)
			if err != nil {
				log.Error(err)
				return fmt.Errorf("parse stacks failed. %w", err)
			}

			// Get absolute path
			absPath, err := filepath.Abs(p)
			if err != nil {
				log.Error(err)
				return fmt.Errorf("project path failed. %w", err)
			}

			project.Stacks = stacks
			project.Path = absPath
			projects = append(projects, project)
		}
		return nil
	})

	return projects, err
}

// IsStack determine whether the given path is Stack directory
func IsStack(path string) bool {
	f, err := os.Stat(path)
	f2, err2 := os.Stat(filepath.Join(path, StackFile))

	if (err == nil && f.IsDir()) && (err2 == nil && f2.Mode().IsRegular()) {
		return true
	}

	return false
}

// GetStackFrom get stack from the given path
func GetStackFrom(path string) (*v1.Stack, error) {
	if !IsStack(path) {
		return nil, ErrNotStackDirectory
	}

	stacks, err := FindAllStacksFrom(path)
	if err != nil {
		return nil, err
	}

	if len(stacks) != 1 {
		return nil, ErrStackNotUnique
	}

	return stacks[0], nil
}

// FindAllStacksFrom find all stacks from the given path
func FindAllStacksFrom(path string) ([]*v1.Stack, error) {
	var stacks []*v1.Stack
	s := sets.NewString()
	_ = filepath.WalkDir(path, func(p string, _ fs.DirEntry, _ error) (err error) {
		if IsStack(p) && !s.Has(p) {
			// Parse stack.yaml
			stack, err := parseStackYamlFile(p)
			if err != nil {
				log.Error(err)
				return nil
			}

			// Get absolute path
			absPath, err := filepath.Abs(p)
			if err != nil {
				log.Error(err)
				return nil
			}

			stack.Path = absPath
			stacks = append(stacks, stack)
		}

		return nil
	})

	return stacks, nil
}

// ParseProjectConfiguration parse the project configuration by the given directory
func parseProjectYamlFile(path string) (*v1.Project, error) {
	var project v1.Project

	err := parseYamlFile(filepath.Join(path, ProjectFile), &project)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

// parseStackConfiguration parse the stack configuration by the given directory
func parseStackYamlFile(path string) (*v1.Stack, error) {
	var stack v1.Stack

	err := parseYamlFile(filepath.Join(path, StackFile), &stack)
	if err != nil {
		return nil, err
	}

	return &stack, nil
}

// Parse yaml data by file name
func parseYamlFile(filename string, target interface{}) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = yamlv3.Unmarshal(content, target)
	if err != nil {
		return err
	}

	return nil
}

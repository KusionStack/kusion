package projectstack

import (
	"io/fs"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/sets"

	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util/yaml"
	"kusionstack.io/kusion/third_party/pulumi/fsutil"
)

// IsProject determine whether the given path is Project directory
func IsProject(path string) bool {
	f, err := os.Stat(path)
	f2, err2 := os.Stat(filepath.Join(path, ProjectFile))

	if (err == nil && f.IsDir()) && (err2 == nil && f2.Mode().IsRegular()) {
		return true
	}

	return false
}

// IsProjectFile determine whether the given path is Project file
func IsProjectFile(path string) bool {
	f, err := os.Stat(path)
	return err == nil && !f.IsDir() && f.Mode().IsRegular() && filepath.Base(path) == ProjectFile
}

// FindProjectPath locates the closest project from the current working directory, or an error if not found.
func FindProjectPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	path, err := FindProjectPathFrom(dir)
	if err != nil {
		return "", err
	}

	return path, nil
}

// FindProjectPathFrom locates the closest project from the given path, searching "upwards" in the directory
// hierarchy.  If no project is found, an empty path is returned.
func FindProjectPathFrom(path string) (string, error) {
	file, err := fsutil.WalkUp(path, IsProjectFile, func(s string) bool {
		return true
	})
	if err != nil {
		return "", err
	}

	return filepath.Dir(file), nil
}

// ParseProjectConfiguration parse the project configuration by the given directory
func ParseProjectConfiguration(path string) (*ProjectConfiguration, error) {
	if !IsProject(path) {
		return nil, ErrNotProjectDirectory
	}

	var config ProjectConfiguration

	err := yaml.ParseYamlFromFile(filepath.Join(path, ProjectFile), &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// FindAllProjects find all projects from the current working directory
func FindAllProjects() ([]*Project, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	projects, err := FindAllProjectsFrom(dir)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

// FindAllProjectsFrom find all project from the given path
func FindAllProjectsFrom(path string) ([]*Project, error) {
	projects := []*Project{}
	s := sets.NewString()
	_ = filepath.WalkDir(path, func(p string, _ fs.DirEntry, _ error) error {
		if IsProject(p) && !s.Has(p) {
			// Parse project configuration
			config, err := ParseProjectConfiguration(p)
			if err != nil {
				log.Error(err)
				return nil
			}

			// Find all stacks
			stacks, err := FindAllStacksFrom(p)
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

			projects = append(projects, NewProject(config, absPath, stacks))
		}

		return nil
	})

	return projects, nil
}

// GetProject get project from the current working directory
func GetProject() (*Project, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	project, err := GetProjectFrom(dir)
	if err != nil {
		return nil, err
	}

	return project, nil
}

// GetProjectFrom get project from the given path
func GetProjectFrom(path string) (*Project, error) {
	if !IsProject(path) {
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

// DetectProjectAndStack try to get stack and project from given path
func DetectProjectAndStack(stackDir string) (project *Project, stack *Stack, err error) {
	stackDir, err = filepath.Abs(stackDir)
	if err != nil {
		return nil, nil, err
	}

	stack, err = GetStackFrom(stackDir)
	if err != nil {
		return nil, nil, err
	}

	projectDir, err := FindProjectPathFrom(stackDir)
	if err != nil {
		return nil, nil, err
	}

	project, err = GetProjectFrom(projectDir)
	if err != nil {
		return nil, nil, err
	}

	return project, stack, nil
}

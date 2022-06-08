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

// IsStack determine whether the given path is Stack directory
func IsStack(path string) bool {
	f, err := os.Stat(path)
	f2, err2 := os.Stat(filepath.Join(path, StackFile))

	if (err == nil && f.IsDir()) && (err2 == nil && f2.Mode().IsRegular()) {
		return true
	}

	return false
}

// IsStackFile determine whether the given path is Stack file
func IsStackFile(path string) bool {
	f, err := os.Stat(path)
	return err == nil && !f.IsDir() && f.Mode().IsRegular() && filepath.Base(path) == StackFile
}

// FindStackPath locates the closest stack from the current working directory, or an error if not found.
func FindStackPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	path, err := FindStackPathFrom(dir)
	if err != nil {
		return "", err
	}

	return path, nil
}

// FindStackPathFrom locates the closest stack from the given path, searching "upwards" in the directory
// hierarchy.  If no stack is found, an empty path is returned.
func FindStackPathFrom(path string) (string, error) {
	file, err := fsutil.WalkUp(path, IsStackFile, func(s string) bool {
		return true
	})
	if err != nil {
		return "", err
	}

	return filepath.Dir(file), nil
}

// ParseStackConfiguration parse the stack configuration by the given directory
func ParseStackConfiguration(path string) (*StackConfiguration, error) {
	if !IsStack(path) {
		return nil, ErrNotStackDirectory
	}

	var stack StackConfiguration

	err := yaml.ParseYamlFromFile(filepath.Join(path, StackFile), &stack)
	if err != nil {
		return nil, err
	}

	return &stack, nil
}

// FindAllStacks find all stacks from the current working directory
func FindAllStacks() ([]*Stack, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	stacks, err := FindAllStacksFrom(dir)
	if err != nil {
		return nil, err
	}

	return stacks, nil
}

// FindAllStacksFrom find all stacks from the given path
func FindAllStacksFrom(path string) ([]*Stack, error) {
	stacks := []*Stack{}
	s := sets.NewString()
	_ = filepath.WalkDir(path, func(p string, _ fs.DirEntry, _ error) (err error) {
		if IsStack(p) && !s.Has(p) {
			// Parse stack configuration
			config, err := ParseStackConfiguration(p)
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

			stacks = append(stacks, NewStack(config, absPath))
		}

		return nil
	})

	return stacks, nil
}

// GetStack get stack from the current working directory
func GetStack() (*Stack, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	stack, err := GetStackFrom(dir)
	if err != nil {
		return nil, err
	}

	return stack, nil
}

// GetStackFrom get stack from the given path
func GetStackFrom(path string) (*Stack, error) {
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

package deps

import (
	"fmt"
	kcl "kusionstack.io/kclvm-go"
	"kusionstack.io/kclvm-go/pkg/tools/list"
	"kusionstack.io/kusion/pkg/projectstack"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type DepsOptions struct {
	workDir string
	Direct  string
	Focus   []string
	Only    string
	Ignore  []string
}

// stringSet is a simple string set implementation by map
type stringSet map[string]bool

// emptyStringSet creates a string set with an empty value list
func emptyStringSet() stringSet {
	return make(stringSet)
}

// add a string value to the stringSet s
func (s stringSet) add(value string) {
	s[value] = true
}

// contains checks if the stringSet s contains certain string value
func (s stringSet) contains(value string) bool {
	_, ok := s[value]
	return ok
}

// toSlice generates a string slice containing all the string values in the stringSet s
func (s stringSet) toSlice() []string {
	var result []string
	for value := range s {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func NewDepsOptions() *DepsOptions {
	return &DepsOptions{}
}

func (o *DepsOptions) Complete(args []string) {
	if len(args) > 0 {
		o.workDir = args[0]
	}

	if o.workDir == "" {
		o.workDir, _ = os.Getwd()
	}
}

func (o *DepsOptions) Validate() error {
	if o.Only != "project" && o.Only != "stack" {
		return fmt.Errorf("invalid output downstream type. supported types: project, stack")
	}

	if o.Direct != "down" && o.Direct != "up" {
		return fmt.Errorf("invalid output direction of the dependency inspection. supported directions: up, down")
	}

	if _, err := os.Stat(o.workDir); err != nil {
		return fmt.Errorf("invalid work dir: %s", err)
	}

	if o.Focus == nil || len(o.Focus) == 0 {
		return fmt.Errorf("invalid focus paths. cannot be empty")
	}

	for _, ignore := range o.Ignore {
		if _, err := os.Stat(filepath.Join(o.workDir, ignore)); err != nil {
			return fmt.Errorf("invalid ignore path. need to be valid relative path from the workdir: %s", err)
		}
	}
	return nil
}

func (o *DepsOptions) Run() (err error) {
	workDir, err := filepath.Abs(o.workDir)
	if err != nil {
		return err
	}
	o.workDir = workDir
	switch o.Direct {
	// list upstream files of the focus files
	case "up":
		depsFiles, err := list.ListUpStreamFiles(o.workDir, &list.DepOptions{Files: o.Focus})
		if err != nil {
			return err
		}
		for _, v := range depsFiles {
			fmt.Println(v)
		}
		return nil
	// list downstream files of the focus files
	case "down":
		var projectOnly bool
		switch o.Only {
		case "project":
			projectOnly = true
		case "stack":
			projectOnly = false
		default:
			return fmt.Errorf("invalid output downstream type. supported types: project, stack")
		}

		// 1. index the paths that should be ignored and the paths that need to list downstream stacks/projects on.
		shouldIgnore := toSet(o.Ignore)
		focusPaths := toSet(o.Focus)

		entrances := emptyStringSet()                // all the entrance files in the work directory
		entranceIndex := map[string][]string{}       // entrance file index to the corresponding projects/stacks
		downstreamProjectOrStack := emptyStringSet() // might be downstream projects or downstream stacks, according to the user option

		// 2. find all projects from work directory. for each project/stack, check if there are files changed in it or collect all the entrance files in them
		// To list downstream stacks/projects, wee need to go through all the entrance files under each project/stack,
		// then filter out the ones that are downstream of the focus files
		var projects []*projectstack.Project
		if projects, err = projectstack.FindAllProjectsFrom(o.workDir); err != nil {
			return err
		}
		for _, project := range projects {
			projectRelative, _ := filepath.Rel(o.workDir, project.GetPath())
			for _, stack := range project.Stacks {
				// 2.1 get the relative path of project/stack
				stackRelative, _ := filepath.Rel(o.workDir, stack.GetPath())
				var stackProjectPath string
				if projectOnly {
					stackProjectPath = projectRelative
				} else {
					stackProjectPath = stackRelative
				}
				// 2.2 in order to save time, cut off the focus paths which are exactly under some stacks/projects directory
				// so that:
				// a. the corresponding stacks/projects can be directly marked as downstream
				// b. those focus paths will be skipped to call the ListDownStreamFiles API
				isDownstream := false
				// iterate all the focus files and check if the dir path is changed, and delete files which is under the dir path
				for f := range focusPaths {
					if strings.HasPrefix(f, stackRelative) {
						// skip those focus paths that are under the stack directory
						delete(focusPaths, f)
						if _, ok := shouldIgnore[f]; !ok {
							isDownstream = true
						}
					}
				}
				if isDownstream {
					// mark the stack/project as downstream
					downstreamProjectOrStack.add(stackProjectPath)
					continue
				}
				// 2.3 collect and index all the entrance files of the stack by loading the settings file
				settingsPath := filepath.Join(stack.GetPath(), projectstack.KclFile)
				opt := kcl.WithSettings(settingsPath)
				if opt.Err != nil {
					// the stack settings is invalid
					return fmt.Errorf("invalid settings file(%s), %v", settingsPath, err)
				}
				for _, entranceFile := range opt.KFilenameList {
					entranceRel, _ := filepath.Rel(o.workDir, entranceFile)
					entranceIndex[entranceRel] = append(entranceIndex[entranceRel], stackProjectPath)
					entrances.add(entranceRel)
				}
			}
		}

		// 3. call the ListDownStreamFiles API
		downstreamFiles, err := kcl.ListDownStreamFiles(o.workDir, &list.DepOptions{
			Files:     entrances.toSlice(),
			UpStreams: focusPaths.toSlice(),
		})
		if err != nil {
			return err
		}

		// 4. for each downstream file, check if it belongs to some stacks/projects, and mark those stacks/projects as downstream
		for _, file := range downstreamFiles {
			if stacksOrProjs, ok := entranceIndex[filepath.Join(o.workDir, file)]; ok {
				for _, v := range stacksOrProjs {
					downstreamProjectOrStack.add(v)
				}
			}
		}

		// 5. output the result
		for name := range downstreamProjectOrStack {
			fmt.Println(name)
		}
		return nil
	default:
		return fmt.Errorf("unsupport diretion")
	}
}

func toSet(list []string) stringSet {
	result := emptyStringSet()
	for _, item := range list {
		result.add(item)
	}
	return result
}

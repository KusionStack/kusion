package deps

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	kcl "kusionstack.io/kclvm-go"
	"kusionstack.io/kclvm-go/pkg/tools/list"

	"kusionstack.io/kusion/pkg/projectstack"
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

// remove a string value from the stringSet s
func (s stringSet) remove(value string) {
	delete(s, value)
}

// contains checks if the stringSet s contains certain string value
func (s stringSet) contains(value string) bool {
	_, ok := s[value]
	return ok
}

// toSlice generates a string slice containing all the string values in the stringSet s
func (s stringSet) toSlice() []string {
	result := make([]string, 0, len(s))
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
		return
	}
	o.workDir = workDir
	switch o.Direct {
	// List upstream files of the focus files
	case "up":
		upstreamFiles, err := list.ListUpStreamFiles(o.workDir, &list.DepOptions{Files: o.Focus})
		if err != nil {
			return err
		}
		for _, f := range upstreamFiles {
			fmt.Println(f)
		}
		return nil
	// List downstream files of the focus files
	case "down":
		// 1. Check which are wanted: downstream projects or downstream stacks
		var projectOnly bool
		switch o.Only {
		case "project":
			projectOnly = true
		case "stack":
			projectOnly = false
		default:
			return fmt.Errorf("invalid output downstream type. supported types: project, stack")
		}

		// 2. Index the paths that should be ignored and the paths that need to list downstream stacks/projects on.
		shouldIgnore := toSet(o.Ignore)
		focusPaths := toSet(o.Focus)

		// 2.1 If a path in the focusPaths should be ignored, remove it from the focusPaths
		for f := range shouldIgnore {
			if focusPaths.contains(f) {
				focusPaths.remove(f)
			}
		}

		if len(focusPaths) == 0 {
			return nil
		}

		// 3. Find all the projects under the workdir
		var projects []*projectstack.Project
		if projects, err = projectstack.FindAllProjectsFrom(workDir); err != nil {
			return
		}

		// 4. Filter the downstream stacks/projects of the focusPaths
		downstreams, err := findDownStreams(o.workDir, projects, focusPaths, shouldIgnore, projectOnly)
		if err != nil {
			return err
		}

		// 5. Output the result
		for name := range downstreams {
			fmt.Println(name)
		}
		return nil
	default:
		return fmt.Errorf("unsupport diretion of deps")
	}
}

// findDownStreams finds all the downstream projects/stacks of the focusPaths in the given workDir.
// DownStreams can be downstream projects or downstream stacks, up to the projectOnly parameter.
// By downstream stacks, it means that either the entrances files of those stacks are downstream files of the focus paths or
// the files under those stacks' directories appear in the focus paths list.
// By downstream projects, it means that each of those projects contains one or more downstream stacks of the focus paths.
//
// # Parameters
//
// The workDir should be a valid absolute path of the root path of the KCL program directory.
// The focusPaths and shouldIgnore should be valid relative file paths under the workDir.
// The value of the projectOnly decides whether the downstream projects(projectOnly is true) or stacks(projectOnly is false) will be filtered.
//
// # Usage Caution
//
// This is a very time-consuming function based on the FindAllProjectsFrom API of kusion and the ListDownStreamFiles API of kcl.
// Do not call this function with high frequency and please ensure at least 10 seconds interval when calling.
func findDownStreams(workDir string, projects []*projectstack.Project, focusPaths, shouldIgnore stringSet, projectOnly bool) (downStreams stringSet, err error) {
	entrances := emptyStringSet()               // all the entrance files in the work directory
	entranceIndex := make(map[string]stringSet) // entrance file index to the corresponding projects/stacks
	downStreams = emptyStringSet()              // might be downstream projects or downstream stacks, according to the projectOnly option

	// 1. For each project/stack, check if there are files changed in it or collect all the entrance files in them
	// To list downstream stacks/projects, wee need to go through all the entrance files under each project/stack,
	// then filter out the ones that are downstream of the focus files
	for _, project := range projects {
		projectRel, _ := filepath.Rel(workDir, project.GetPath())
		for _, stack := range project.Stacks {
			// 1.1 Get the relative path of project/stack
			stackRel, _ := filepath.Rel(workDir, stack.GetPath())
			var stackProjectPath string
			if projectOnly {
				stackProjectPath = projectRel
			} else {
				stackProjectPath = stackRel
			}
			// 1.2 In order to save time, cut off the focus paths which are exactly under some stacks/projects directory,
			// so that:
			// a. The corresponding stacks/projects can be directly marked as downstream
			// b. Those focus paths will be skipped to call the ListDownStreamFiles API
			isDownstream := false
			// Iterate all the focus files and check if there appear some files under that stack, and delete those files from the focus paths
			for f := range focusPaths {
				if strings.HasPrefix(f, stackRel) {
					// Skip those focus paths that are under the stack directory
					focusPaths.remove(f)
					if !shouldIgnore.contains(f) {
						isDownstream = true
					}
				}
			}
			if isDownstream {
				// Mark the stack/project as downstream
				downStreams.add(stackProjectPath)
				continue
			}
			// 1.3 Collect and index all the entrance files of the stack by loading the settings file
			settingsPath := filepath.Join(stack.GetPath(), projectstack.KclFile)
			opt := kcl.WithSettings(settingsPath)
			if opt.Err != nil {
				// The stack settings is invalid
				err = fmt.Errorf("invalid settings file(%s), %v", settingsPath, err)
				return
			}
			var entranceRels []string
			for _, entranceFile := range opt.KFilenameList {
				entranceRel, _ := filepath.Rel(workDir, entranceFile)
				if shouldIgnore.contains(entranceRel) {
					// Skip recording entrance files that should be ignored
					continue
				}
				entranceRels = append(entranceRels, entranceRel)
				if focusPaths.contains(entranceRel) {
					// As long as one entrance file appears in the focus paths is enough to mark the stack/project as downstream
					isDownstream = true
				}
			}
			if isDownstream {
				downStreams.add(stackProjectPath)
				continue
			}
			for _, entranceRel := range entranceRels {
				entrances.add(entranceRel)
				if _, ok := entranceIndex[entranceRel]; !ok {
					entranceIndex[entranceRel] = emptyStringSet()
				}
				entranceIndex[entranceRel].add(stackProjectPath)
			}
		}
	}

	// 2. Call the ListDownStreamFiles API
	downstreamFiles, err := kcl.ListDownStreamFiles(workDir, &list.DepOptions{
		Files:     entrances.toSlice(),
		UpStreams: focusPaths.toSlice(),
	})
	if err != nil {
		return
	}

	// 3. For each downstream file, check if it belongs to some stacks/projects, and mark those stacks/projects as downstream
	for _, file := range downstreamFiles {
		if stacksOrProjs, ok := entranceIndex[file]; ok {
			for v := range stacksOrProjs {
				downStreams.add(v)
			}
		}
	}
	return
}

func toSet(list []string) stringSet {
	result := emptyStringSet()
	for _, item := range list {
		result.add(item)
	}
	return result
}

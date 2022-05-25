package deps

import (
	"fmt"
	"os"
	"path/filepath"

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

	for _, focus := range o.Focus {
		if _, err := os.Stat(filepath.Join(o.workDir, focus)); err != nil {
			return fmt.Errorf("invalid focus path. need to be valid relative path from the workdir: %s", err)
		}
	}

	for _, ignore := range o.Ignore {
		if _, err := os.Stat(filepath.Join(o.workDir, ignore)); err != nil {
			return fmt.Errorf("invalid ignore path. need to be valid relative path from the workdir: %s", err)
		}
	}
	return nil
}

func (o *DepsOptions) Run() error {
	workDir, err := filepath.Abs(o.workDir)
	if err != nil {
		return err
	}
	o.workDir = workDir
	switch o.Direct {
	case "up":
		depsFiles, err := list.ListUpStreamFiles(o.workDir, &list.DepOption{Files: o.Focus})
		if err != nil {
			return err
		}
		for _, v := range depsFiles {
			fmt.Println(v)
		}
		return nil
	case "down":
		projects, err := projectstack.FindAllProjectsFrom(o.workDir)
		if err != nil {
			return err
		}
		file2StackMap := map[string][]string{}
		file2ProjMap := map[string][]string{}
		entranceFiles := []string{}
		ignoreMap := map[string]bool{}
		for _, ignore := range o.Ignore {
			ignoreMap[ignore] = true
		}
		for _, project := range projects {
			relProjPath, err := filepath.Rel(o.workDir, project.GetPath())
			if err != nil {
				return err
			}
			for _, stack := range project.Stacks {
				relStackPath, err := filepath.Rel(o.workDir, stack.GetPath())
				if err != nil {
					return err
				}
				opt := kcl.WithSettings(filepath.Join(stack.GetPath(), projectstack.KclFile))
				for _, entranceFile := range opt.KFilenameList {
					relPath, err := filepath.Rel(o.workDir, entranceFile)
					if err != nil {
						return err
					}
					if _, ok := ignoreMap[relPath]; ok {
						continue
					}
					file2StackMap[relPath] = append(file2StackMap[relPath], relStackPath)
					file2ProjMap[relPath] = append(file2ProjMap[relPath], relProjPath)
					entranceFiles = append(entranceFiles, relPath)
				}
				sFiles, err := listFiles(stack.GetPath(), true)
				if err != nil {
					return err
				}
				for _, file := range sFiles {
					rel, _ := filepath.Rel(o.workDir, file)
					if _, ok := file2StackMap[rel]; !ok {
						file2StackMap[rel] = append(file2StackMap[rel], relStackPath)
						file2ProjMap[rel] = append(file2ProjMap[rel], relProjPath)
					}
				}
			}
		}
		affectedFiles, err := kcl.ListDownStreamFiles(o.workDir, &list.DepOption{
			Files:        entranceFiles,
			ChangedPaths: o.Focus,
		})
		if err != nil {
			return err
		}
		var fileMap map[string][]string
		switch o.Only {
		case "project":
			fileMap = file2ProjMap
		case "stack":
			fileMap = file2StackMap
		default:
			return fmt.Errorf("invalid output downstream type. supported types: project, stack")
		}
		affected := map[string]bool{}
		for _, affect := range affectedFiles {
			// filter
			if stacks, ok := fileMap[affect]; ok {
				for _, stack := range stacks {
					affected[stack] = true
				}
			}
		}
		for name := range affected {
			fmt.Println(name)
		}
		return nil
	default:
		return fmt.Errorf("unsupport diretion")
	}
}

func listFiles(root string, resursive bool) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	files := []string{}
	for _, file := range entries {
		if !file.IsDir() {
			files = append(files, filepath.Join(root, file.Name()))
		} else if resursive {
			subFiles, err := listFiles(filepath.Join(root, file.Name()), resursive)
			if err != nil {
				return nil, err
			}
			files = append(files, subFiles...)
		}
	}
	return files, nil
}

package ls

import (
	"fmt"
	"time"

	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"

	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/util/gitutil"
	"kusionstack.io/kusion/pkg/util/json"
	"kusionstack.io/kusion/pkg/util/yaml"
)

type lsReport struct {
	WorkDir     string                  `json:"workDir" yaml:"workDir"` // Such as "/Users/admin/workspace/Konfig/appops/http-echo"
	WorkGitInfo *gitutil.GitInfo        `json:"workGitInfo,omitempty" yaml:"workGitInfo,omitempty"`
	Time        string                  `json:"time,omitempty" yaml:"time,omitempty"` // Such as "2021-10-20 18:24:03"
	Projects    []*projectstack.Project `json:"projects" yaml:"projects"`
}

func NewLsReport(workDir string, projects []*projectstack.Project) *lsReport {
	gitInfo, err := gitutil.NewGitInfoFrom(workDir)
	if err != nil {
		log.Warnf("Failed to get git info as [%v]", err)
	}

	return &lsReport{
		WorkDir:     workDir,
		WorkGitInfo: gitInfo,
		Time:        time.Now().Format("2006-01-02 15:04:05"),
		Projects:    projects,
	}
}

func (r *lsReport) Human() (string, error) {
	// Convert []*Project to []NameAndPath
	projectNameAndPaths := []NameAndPath{}
	for _, project := range r.Projects {
		projectNameAndPaths = append(projectNameAndPaths, NameAndPath(project))
	}

	// Prompt project
	projectI, err := promptProjectOrStack(projectNameAndPaths, "project")
	if err != nil {
		return "", err
	}

	project, ok := projectI.(*projectstack.Project)
	if !ok {
		return "", fmt.Errorf("invalid convert for project")
	}

	fmt.Println(project.TableReport())

	// Convert []*Stack to []NameAndPath
	stackNameAndPaths := []NameAndPath{}
	for _, stack := range project.Stacks {
		stackNameAndPaths = append(stackNameAndPaths, NameAndPath(stack))
	}

	// Prompt stack
	stackI, err := promptProjectOrStack(stackNameAndPaths, "stack")
	if err != nil {
		return "", err
	}

	stack, ok := stackI.(*projectstack.Stack)
	if !ok {
		return "", fmt.Errorf("invalid convert for stack")
	}

	return stack.TableReport(), nil
}

func (r *lsReport) Tree() (string, error) {
	// Fill tree content
	leveledList := pterm.LeveledList{}

	if r.Projects != nil {
		for _, project := range r.Projects {
			leveledList = append(leveledList, pterm.LeveledListItem{Level: 0, Text: project.Name})

			if project.Stacks != nil {
				for _, stack := range project.Stacks {
					leveledList = append(leveledList, pterm.LeveledListItem{Level: 1, Text: stack.Name})
				}
			}
		}
	}

	// Generate tree from LeveledList.
	root := putils.TreeFromLeveledList(leveledList)

	// Render TreePrinter
	tree, err := pterm.DefaultTree.WithRoot(root).Srender()
	if err != nil {
		log.Warnf("Failed to render tree as [%v]", err)
		return "", err
	}

	return tree, nil
}

func (r *lsReport) YAML() (string, error) {
	return yaml.MergeToOneYAML(r), nil
}

func (r *lsReport) JSON() (string, error) {
	return json.MustMarshal2PrettyString(r), nil
}

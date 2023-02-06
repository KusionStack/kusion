package ls

import (
	"fmt"
	"os"

	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/util/pretty"
)

type LsOptions struct {
	workDir      string
	OutputFormat string
	Level        int
}

func NewLsOptions() *LsOptions {
	return &LsOptions{}
}

func (o *LsOptions) Complete(args []string) {
	if len(args) > 0 {
		o.workDir = args[0]
	}

	if o.workDir == "" {
		o.workDir, _ = os.Getwd()
	}
}

func (o *LsOptions) Validate() error {
	if o.OutputFormat != "json" && o.OutputFormat != "yaml" && o.OutputFormat != "tree" && o.OutputFormat != "human" {
		return fmt.Errorf("invalid output format. supported formats: json, yaml, tree, human")
	}

	if !(0 <= o.Level && o.Level <= 2) {
		return fmt.Errorf("invalid output level: must between 0-2")
	}

	if _, err := os.Stat(o.workDir); err != nil {
		return fmt.Errorf("invalid work dir: %s", err)
	}

	return nil
}

func filterProjectsByLevel(projects []*projectstack.Project, level int) []*projectstack.Project {
	newProjects := []*projectstack.Project{}

	if level > 0 {
		for _, project := range projects {
			newProject := projectstack.NewProject(&project.ProjectConfiguration, project.Path, []*projectstack.Stack{})
			newProjects = append(newProjects, newProject)

			if level > 1 {
				newProject.Stacks = append(newProject.Stacks, project.Stacks...)
			}
		}
	}

	return newProjects
}

func (o *LsOptions) Run() error {
	// Parse all projects and stacks
	sp, err := pretty.SpinnerT.WithRemoveWhenDone(true).Start("Parsing projects and stacks ...")
	if err != nil {
		return err
	}

	projects, err := projectstack.FindAllProjectsFrom(o.workDir)
	if err != nil {
		sp.Fail(err.Error())
		return err
	}

	sp.Success()

	// Filter projects by level
	projects = filterProjectsByLevel(projects, o.Level)

	// Build ls report
	r := NewLsReport(o.workDir, projects)

	// Output
	var output string

	switch o.OutputFormat {
	case "json":
		output, err = r.JSON()
	case "yaml":
		output, err = r.YAML()
	case "tree":
		output, err = r.Tree()
	case "human":
		output, err = r.Human()
	default:
		return fmt.Errorf("invalid output format. supported formats: json, yaml, tree, human")
	}

	if err != nil {
		return err
	}

	fmt.Println(output)

	return nil
}

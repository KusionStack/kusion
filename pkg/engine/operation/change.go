package operation

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pterm/pterm"

	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/util/diff"
	"kusionstack.io/kusion/pkg/util/pretty"
	"kusionstack.io/kusion/pkg/util/yaml"
	"kusionstack.io/kusion/third_party/dyff"
)

type ChangeStep struct {
	ID     string      // the resource id
	Action ActionType  // the operation performed by this step.
	Old    interface{} // the state of the resource before performing this step.
	New    interface{} // the state of the resource after performing this step.
}

func (cs *ChangeStep) Diff() (string, error) {
	// Generate diff report
	diffReport, err := diffToReport(cs.Old, cs.New)
	if err != nil {
		log.Errorf("failed to compute diff with ChangeStep ID: %s", cs.ID)
		return "", err
	}

	reportString, err := diff.ToReportString(*diffReport)
	if err != nil {
		log.Warn("diff to string error: %v", err)
		return "", err
	}

	buf := bytes.NewBufferString("")

	if len(cs.ID) != 0 {
		buf.WriteString(pretty.GreenBold("ID: "))
		buf.WriteString(pretty.Green("%s\n", cs.ID))
	}
	if cs.Action != Undefined {
		buf.WriteString(pretty.GreenBold("Plan: "))
		buf.WriteString(pterm.Sprintf("%s\n", cs.Action.PrettyString()))
	}
	buf.WriteString(pretty.GreenBold("Diff: "))
	if len(strings.TrimSpace(reportString)) == 0 && cs.Action == UnChange {
		buf.WriteString(pretty.Gray("<EMPTY>"))
	} else {
		buf.WriteString("\n" + strings.TrimSpace(reportString))
	}
	buf.WriteString("\n")
	return buf.String(), nil
}

func NewChangeStep(id string, op ActionType, oldData, newData interface{}) *ChangeStep {
	return &ChangeStep{
		ID:     id,
		Action: op,
		Old:    oldData,
		New:    newData,
	}
}

type ChangeStepFilterFunc func(*ChangeStep) bool

var (
	CreateChangeStepFilter   = func(c *ChangeStep) bool { return c.Action == Create }
	UpdateChangeStepFilter   = func(c *ChangeStep) bool { return c.Action == Update }
	DeleteChangeStepFilter   = func(c *ChangeStep) bool { return c.Action == Delete }
	UnChangeChangeStepFilter = func(c *ChangeStep) bool { return c.Action == UnChange }
)

type Changes struct {
	ChangeSteps map[string]*ChangeStep
	project     *projectstack.Project // the project of current changes
	stack       *projectstack.Stack   // the stack of current changes
}

func NewChanges(p *projectstack.Project, s *projectstack.Stack, steps map[string]*ChangeStep) *Changes {
	return &Changes{
		ChangeSteps: steps,
		project:     p,
		stack:       s,
	}
}

func (p *Changes) Get(key string) *ChangeStep {
	return p.ChangeSteps[key]
}

func (p *Changes) Values(filters ...ChangeStepFilterFunc) []*ChangeStep {
	result := []*ChangeStep{}

	for _, v := range p.ChangeSteps {
		// Deal filters
		var i int
		for i = 0; i < len(filters); i++ {
			if !filters[i](v) {
				break
			}
		}
		if i < len(filters) {
			continue
		}

		// Append item to result
		result = append(result, v)
	}

	return result
}

func (p *Changes) Stack() *projectstack.Stack {
	return p.stack
}

func (p *Changes) Project() *projectstack.Project {
	return p.project
}

func (p *Changes) Diffs() string {
	buf := bytes.NewBufferString("")

	for _, step := range p.ChangeSteps {
		// Generate diff report
		diffString, err := step.Diff()
		if err != nil {
			log.Errorf("failed to generate diff string with ChangeStep ID: %s", step.ID)
			continue
		}

		buf.WriteString(diffString)
	}
	return buf.String()
}

func (p *Changes) Summary() {
	// Create a fork of the default table, fill it with data and print it.
	// Data can also be generated and inserted later.
	tableHeader := []string{fmt.Sprintf("Stack: %s", p.stack.Name), "ID", "Action"}
	tableData := pterm.TableData{tableHeader}

	for i, step := range p.Values() {
		itemPrefix := " * ├─"
		if i == len(p.ChangeSteps)-1 {
			itemPrefix = " * └─"
		}

		tableData = append(tableData, []string{itemPrefix, step.ID, step.Action.String()})
	}

	pterm.DefaultTable.WithHasHeader().
		// WithBoxed(true).
		WithHeaderStyle(&pterm.ThemeDefault.TableHeaderStyle).
		WithRightAlignment(true).
		WithSeparator("  ").
		WithData(tableData).
		Render()
	pterm.Println() // Blank line
}

func (p *Changes) PromptDetails() (string, error) {
	// Prepare the selects
	options := []string{"all"}
	optionMaps := map[string]string{"all": "all"}

	for _, cs := range p.ChangeSteps {
		humanKeyAndOp := pterm.Sprintf("%s %s", cs.ID, pretty.Gray(cs.Action.String()))
		options = append(options, humanKeyAndOp)
		optionMaps[humanKeyAndOp] = cs.ID
	}

	options = append(options, "cancel")

	// Build prompt
	prompt := &survey.Select{
		Message: `Which diff detail do you want to see?`,
		Options: options,
	}

	// Run prompt
	var input string
	err := survey.AskOne(prompt, &input)
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return "", err
	}
	return optionMaps[input], nil
}

func (p *Changes) OutputDiff(target string) {
	switch target {
	case "all":
		fmt.Println(p.Diffs())
	default:
		rinID := target
		if cs, ok := p.ChangeSteps[rinID]; ok {
			diffString, err := cs.Diff()
			if err != nil {
				log.Error("failed to output specify diff with rinID: %s, err: %v", rinID, err)
			}

			fmt.Println(diffString)
		}
	}
}

func buildResourceStateMap(rs []*states.ResourceState) map[string]*states.ResourceState {
	rMap := make(map[string]*states.ResourceState)
	if len(rs) == 0 {
		return rMap
	}
	for _, r := range rs {
		if key := r.ResourceKey(); key != "" {
			rMap[key] = r
		}
	}

	return rMap
}

func diffToReport(oldData, newData interface{}) (*dyff.Report, error) {
	from, err := LoadFile(yaml.MergeToOneYAML(oldData), "Old item")
	if err != nil {
		return nil, err
	}

	to, err := LoadFile(yaml.MergeToOneYAML(newData), "New item")
	if err != nil {
		return nil, err
	}

	report, err := dyff.CompareInputFiles(from, to, dyff.IgnoreOrderChanges(true))
	if err != nil {
		return nil, err
	}
	return &report, nil
}

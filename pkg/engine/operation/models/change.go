package models

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pterm/pterm"

	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/util/diff"
	"kusionstack.io/kusion/pkg/util/pretty"
)

type ChangeStep struct {
	// the resource id
	ID string `json:"id,omitempty" yaml:"id,omitempty"`
	// the operation performed by this step
	Action ActionType `json:"action,omitempty" yaml:"action,omitempty"`
	// old data
	From interface{} `json:"from,omitempty" yaml:"from,omitempty"`
	// new data
	To interface{} `json:"to,omitempty" yaml:"to,omitempty"`
}

// Diff compares objects(from and to) which stores in ChangeStep,
// and return a human-readable string report.
func (cs *ChangeStep) Diff() (string, error) {
	// Generate diff report
	diffReport, err := diff.ToReport(cs.From, cs.To)
	if err != nil {
		log.Errorf("failed to compute diff with ChangeStep ID: %s", cs.ID)
		return "", err
	}

	reportString, err := diff.ToHumanString(diff.NewHumanReport(diffReport))
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
	if len(strings.TrimSpace(reportString)) == 0 && cs.Action == UnChanged {
		buf.WriteString(pretty.Gray("<EMPTY>"))
	} else {
		buf.WriteString("\n" + strings.TrimSpace(reportString))
	}
	buf.WriteString("\n")
	return buf.String(), nil
}

func NewChangeStep(id string, op ActionType, from, to interface{}) *ChangeStep {
	return &ChangeStep{
		ID:     id,
		Action: op,
		From:   from,
		To:     to,
	}
}

type ChangeStepFilterFunc func(*ChangeStep) bool

var (
	CreateChangeStepFilter   = func(c *ChangeStep) bool { return c.Action == Create }
	UpdateChangeStepFilter   = func(c *ChangeStep) bool { return c.Action == Update }
	DeleteChangeStepFilter   = func(c *ChangeStep) bool { return c.Action == Delete }
	UnChangeChangeStepFilter = func(c *ChangeStep) bool { return c.Action == UnChanged }
)

type Changes struct {
	*ChangeOrder `json:",inline" yaml:",inline"`

	project *projectstack.Project // the project of current changes
	stack   *projectstack.Stack   // the stack of current changes
}

type ChangeOrder struct {
	StepKeys    []string               `json:"stepKeys,omitempty" yaml:"stepKeys,omitempty"`
	ChangeSteps map[string]*ChangeStep `json:"changeSteps,omitempty" yaml:"changeSteps,omitempty"`
}

func NewChanges(p *projectstack.Project, s *projectstack.Stack, order *ChangeOrder) *Changes {
	return &Changes{
		ChangeOrder: order,
		project:     p,
		stack:       s,
	}
}

func (o *ChangeOrder) Get(key string) *ChangeStep {
	return o.ChangeSteps[key]
}

func (o *ChangeOrder) Values(filters ...ChangeStepFilterFunc) []*ChangeStep {
	result := []*ChangeStep{}

	for _, key := range o.StepKeys {
		v := o.ChangeSteps[key]
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

func (o *ChangeOrder) Diffs() string {
	buf := bytes.NewBufferString("")

	for _, key := range o.StepKeys {
		step := o.ChangeSteps[key]
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

func (p *Changes) AllUnChange() bool {
	for _, v := range p.ChangeSteps {
		if v.Action != UnChanged {
			return false
		}
	}

	return true
}

func (p *Changes) Summary(writer io.Writer) {
	// Create a fork of the default table, fill it with data and print it.
	// Data can also be generated and inserted later.
	tableHeader := []string{fmt.Sprintf("Stack: %s", p.stack.Name), "ID", "Action"}
	tableData := pterm.TableData{tableHeader}

	for i, step := range p.Values() {
		itemPrefix := " * ├─"
		if i == len(p.StepKeys)-1 {
			itemPrefix = " * └─"
		}

		tableData = append(tableData, []string{itemPrefix, step.ID, step.Action.String()})
	}

	pterm.DefaultTable.WithHasHeader().
		// WithBoxed(true).
		WithHeaderStyle(&pterm.ThemeDefault.TableHeaderStyle).
		WithLeftAlignment(true).
		WithSeparator("  ").
		WithData(tableData).
		WithWriter(writer).
		Render()
	pterm.Println() // Blank line
}

func (o *ChangeOrder) PromptDetails() (string, error) {
	// Prepare the selects
	options := []string{"all"}
	optionMaps := map[string]string{"all": "all"}

	for _, key := range o.StepKeys {
		cs := o.ChangeSteps[key]
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

func (o *ChangeOrder) OutputDiff(target string) {
	switch target {
	case "all":
		fmt.Println(o.Diffs())
	default:
		rinID := target
		if cs, ok := o.ChangeSteps[rinID]; ok {
			diffString, err := cs.Diff()
			if err != nil {
				log.Error("failed to output specify diff with rinID: %s, err: %v", rinID, err)
			}

			fmt.Println(diffString)
		}
	}
}

func buildResourceStateMap(rs []*models.Resource) map[string]*models.Resource {
	rMap := make(map[string]*models.Resource)
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

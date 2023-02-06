package ls

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pterm/pterm"

	"kusionstack.io/kusion/pkg/util/pretty"
)

func buildDisplayName(name, path string, noStyle bool) string {
	if noStyle {
		return name + " " + path
	} else {
		return pterm.Sprintf("%s %s", pretty.GreenBold(name), pretty.Gray(path))
	}
}

// commonSearcher is a common searcher.
func commonSearcher(content, input string) bool {
	if strings.Contains(input, " ") {
		for _, key := range strings.Split(input, " ") {
			key = strings.TrimSpace(key)
			if key != "" {
				if !strings.Contains(content, key) {
					return false
				}
			}
		}

		return true
	}

	if strings.Contains(content, input) {
		return true
	}

	return false
}

type NameAndPath interface {
	GetName() string
	GetPath() string
}

// promptProject prompts project.
func promptProjectOrStack(items []NameAndPath, promptType string) (interface{}, error) {
	// Build the name and path array
	displayNameAndPaths := []string{}

	for _, item := range items {
		displayNameAndPaths = append(displayNameAndPaths, buildDisplayName(item.GetName(), item.GetPath(), false))
	}

	// Build the prompt
	prompt := &survey.Select{
		Message:  fmt.Sprintf("Enter keyword(s) for search %s:", promptType),
		Options:  displayNameAndPaths,
		PageSize: 10,
		Filter: func(filter string, value string, index int) bool {
			item := items[index]
			return commonSearcher(buildDisplayName(item.GetName(), item.GetPath(), true), filter)
		},
	}

	// Prompt
	var selectedNameAndPathColorful string

	err := survey.AskOne(prompt, &selectedNameAndPathColorful, survey.WithIcons(func(icons *survey.IconSet) {
		icons.Question.Text = "🔍"
	}))
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return nil, err
	}

	// Return the selected item
	selectedProjectAndPath := pterm.RemoveColorFromString(selectedNameAndPathColorful)
	temp := strings.Split(selectedProjectAndPath, " ")
	selectedName, selectedPath := temp[0], temp[1]

	for _, item := range items {
		if item.GetName() == selectedName && item.GetPath() == selectedPath {
			return item, nil
		}
	}

	return nil, fmt.Errorf("selected project/stack is not found")
}

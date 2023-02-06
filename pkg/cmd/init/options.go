package init

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pterm/pterm"

	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/scaffold"
)

type InitOptions struct {
	TemplateNameOrURL string
	Online            bool
	ProjectName       string
	Force             bool
	Yes               bool
	CustomParamsJSON  string
}

func NewInitOptions() *InitOptions {
	return &InitOptions{}
}

func (o *InitOptions) Complete(args []string) error {
	if o.Online { // use online templates, official link or user-specified link
		if len(args) > 0 {
			// user-specified link
			o.TemplateNameOrURL = args[0]
		}
	} else { // use offline templates, internal templates or user-specified local dir
		if len(args) > 0 {
			// user-specified local dir
			o.TemplateNameOrURL = args[0]
		} else {
			// use internal templates
			internalTemplateDir, err := scaffold.GetTemplateDir(scaffold.InternalTemplateDir)
			if err != nil {
				return err
			}
			o.TemplateNameOrURL = internalTemplateDir
		}
	}
	return nil
}

func (o *InitOptions) Validate() error {
	if o.Online {
		return nil
	}
	// offline mode may need to generate templates
	internalTemplateDir, err := scaffold.GetTemplateDir(scaffold.InternalTemplateDir)
	if err != nil {
		return err
	}
	// gen internal templates first before using it
	if internalTemplateDir == o.TemplateNameOrURL {
		_, err := os.Stat(o.TemplateNameOrURL)
		if os.IsNotExist(err) {
			return scaffold.GenInternalTemplates()
		}
	}
	return nil
}

func (o *InitOptions) Run() error {
	// Retrieve the template repo.
	repo, err := scaffold.RetrieveTemplates(o.TemplateNameOrURL, o.Online)
	if err != nil {
		return err
	}
	defer func() {
		if err := repo.Delete(); err != nil {
			log.Warnf("Explicitly ignoring and discarding error: %v", err)
		}
	}()

	// List the templates from the repo.
	templates, err := repo.Templates()
	if err != nil {
		return err
	}
	if len(templates) == 0 {
		return errors.New("no templates")
	}

	// Choose template
	var template scaffold.Template
	if template, err = chooseTemplate(templates); err != nil {
		return err
	}

	// Parse customParams if not empty
	var tc scaffold.TemplateConfig
	if o.CustomParamsJSON != "" {
		if err := json.Unmarshal([]byte(o.CustomParamsJSON), &tc); err != nil {
			return err
		}
		if tc.ProjectName == "" {
			return errors.New("`ProjectName` is missing in custom params")
		}

		o.ProjectName = tc.ProjectName
		pterm.Bold.Println("Use custom params to render template...")
	}

	// Show instructions, if we're going to use interactive mode
	if !o.Yes && o.CustomParamsJSON == "" {
		pterm.Println("This command will walk you through creating a new kusion project.")
		pterm.Println()
		pterm.Printfln("Enter a value or leave blank to accept the (default), and press %s.",
			pterm.Cyan("<ENTER>"))
		pterm.Printfln("Press %s at any time to quit.", pterm.Cyan("^C"))
		pterm.Println()
		pterm.Bold.Println("Project Config:")
	}

	// o.ProjectName is used to make root directory
	if o.ProjectName != "" {
		if err := scaffold.ValidateProjectName(o.ProjectName); err != nil {
			return fmt.Errorf("'%s' is not a valid project name as [%v]", o.ProjectName, err)
		}
	} else {
		defaultName := template.ProjectName
		if defaultName == "" {
			defaultName = template.Name
		}
		if !o.Yes {
			o.ProjectName, err = promptValue("Project Name", "ProjectName is a required fully qualified name", defaultName, scaffold.ValidateProjectName)
			if err != nil {
				return err
			}
		} else {
			o.ProjectName = defaultName
		}
	}

	if o.CustomParamsJSON == "" {
		// Prompt project configs from kusion.yaml
		tc.ProjectConfig = make(map[string]interface{})
		for _, f := range template.ProjectFields {
			tc.ProjectConfig[f.Name] = f.Default
			// We don't prompt non-primitive types, such as: array and struct
			if !f.Type.IsPrimitive() || o.Yes {
				continue
			}
			// Prompt and restore actual value
			actual, err := promptAndRestore(f)
			if err != nil {
				return err
			}
			tc.ProjectConfig[f.Name] = actual
		}

		// Prompt stack configs from kusion.yaml
		tc.StacksConfig = make(map[string]map[string]interface{})
		for _, stack := range template.StackTemplates {
			if !o.Yes {
				pterm.Bold.Printfln("Stack Config: %s", pterm.Cyan(stack.Name))
			}
			configs := make(map[string]interface{})
			for _, f := range stack.Fields {
				configs[f.Name] = f.Default
				// We don't prompt non-primitive types, such as: array and struct
				if !f.Type.IsPrimitive() || o.Yes {
					continue
				}
				// Prompt and restore actual value
				actual, err := promptAndRestore(f)
				if err != nil {
					return err
				}
				configs[f.Name] = actual
			}
			tc.StacksConfig[stack.Name] = configs
		}
	}

	// Get the current working directory.
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting the working directory: %w", err)
	}
	// Make dest directory with project name
	desDir := filepath.Join(cwd, o.ProjectName)

	// Actually copy the files.
	if err = scaffold.RenderLocalTemplate(template.Dir, desDir, o.Force, &tc); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("template '%s' not found: %w", template.Name, err)
		}
		return err
	}

	fmt.Printf("Created project '%s'\n", o.ProjectName)

	return nil
}

// chooseTemplate will prompt the user to choose amongst the available templates.
func chooseTemplate(templates []scaffold.Template) (scaffold.Template, error) {
	const chooseTemplateErr = "no template selected; please use `kusion init` to choose one"

	options, optionToTemplateMap := templatesToOptionArrayAndMap(templates)
	message := pterm.Cyan("Please choose a template:")
	prompt := &survey.Select{
		Message:  message,
		PageSize: 10,
		Options:  options,
	}

	var selectedOption scaffold.Template
	var has bool
	var option string

	for {
		err := survey.AskOne(prompt, &option)
		if err != nil {
			return scaffold.Template{}, errors.New(chooseTemplateErr)
		}
		selectedOption, has = optionToTemplateMap[option]
		if has {
			break
		}
	}

	return selectedOption, nil
}

// templatesToOptionArrayAndMap returns an array of option strings and a map of option strings to templates.
// Each option string is made up of the template name and description with some padding in between.
func templatesToOptionArrayAndMap(templates []scaffold.Template) ([]string, map[string]scaffold.Template) {
	// Find the longest name length. Used to add padding between the name and description.
	maxNameLength := 0
	for _, template := range templates {
		if len(template.Name) > maxNameLength {
			maxNameLength = len(template.Name)
		}
	}

	// Build the array and map.
	options := []string{}
	nameToTemplateMap := make(map[string]scaffold.Template)
	for _, template := range templates {
		// Create the option string that combines the name, padding, and description.
		option := fmt.Sprintf(fmt.Sprintf("%%%ds    %%s", -maxNameLength), template.Name, template.Description)

		// Add it to the array and map.
		options = append(options, option)
		nameToTemplateMap[option] = template
	}
	sort.Strings(options)

	return options, nameToTemplateMap
}

// promptAndRestore will prompt f.Value first and restore its actual value based on f.Type
func promptAndRestore(f *scaffold.FieldTemplate) (interface{}, error) {
	// Prompt always return string value, must restore f type
	input, err := promptValue(f.Name, f.Description, fmt.Sprintf("%v", f.Default), nil)
	if err != nil {
		return nil, err
	}
	// Restore f type
	actual, err := f.RestoreActualValue(input)
	if err != nil {
		return nil, err
	}
	return actual, nil
}

func promptValue(valueType string, description string, defaultValue string, isValidFn func(value string) error) (value string, err error) {
	prompt := &survey.Input{
		Message: fmt.Sprintf("%s:", valueType),
		Default: defaultValue,
		Help:    description,
	}

	for {
		// you can pass multiple validators here and survey will make sure each one passes
		err = survey.AskOne(prompt, &value)
		if err != nil {
			return "", err
		}

		// ensure user input is valid
		if isValidFn != nil {
			if validationError := isValidFn(value); validationError != nil {
				// If validation failed, let the user know. If interactive, we will print the error and
				// prompt the user again
				fmt.Printf("Sorry, '%s' is not a valid %s. %s.\n", value, valueType, validationError)
				continue
			}
		}

		break
	}
	return value, nil
}

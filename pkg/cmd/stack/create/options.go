package create

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"kusionstack.io/kusion/pkg/cmd/stack/util"
)

var ErrNotOneArg = errors.New("only one argument is accepted")

type Options struct {
	Name     string
	StackDir string
	Flags
}

type Flags struct {
	ProjectDir string
	CopyFrom   string
}

func NewOptions() *Options {
	return &Options{
		Flags: Flags{
			ProjectDir: "",
			CopyFrom:   "",
		},
	}
}

func (o *Options) Complete(args []string) error {
	if len(args) != 1 {
		return ErrNotOneArg
	}

	o.Name = args[0]

	if o.ProjectDir == "" {
		projectDir, err := os.Getwd()
		if err != nil {
			return err
		}
		o.ProjectDir = projectDir
	} else {
		// Use the absolute path of the target project directory.
		absTargetDir, err := filepath.Abs(o.ProjectDir)
		if err != nil {
			return err
		}
		o.ProjectDir = absTargetDir
	}

	o.StackDir = filepath.Join(o.ProjectDir, o.Name)

	if o.CopyFrom != "" {
		o.CopyFrom = filepath.Join(o.ProjectDir, o.CopyFrom)
	}

	return nil
}

func (o *Options) Validate() error {
	if err := util.ValidateStackName(o.Name); err != nil {
		return err
	}

	if err := util.ValidateProjectDir(o.ProjectDir); err != nil {
		return err
	}

	if err := util.ValidateStackDir(o.StackDir); err != nil {
		return err
	}

	if o.CopyFrom != "" {
		return util.ValidateRefStackDir(o.CopyFrom)
	}

	return nil
}

func (o *Options) Run() error {
	if o.CopyFrom == "" {
		stackYAMLPath := filepath.Join(o.StackDir, util.StackYAMLFile)
		stackYAMLContent := fmt.Sprintf(util.StackYAMLTemplate, o.Name)
		if err := os.WriteFile(stackYAMLPath, []byte(stackYAMLContent), 0o644); err != nil {
			return err
		}

		kclModPath := filepath.Join(o.StackDir, util.KCLModFile)
		kclModContent := fmt.Sprint(util.KCLModFileTemplate)
		if err := os.WriteFile(kclModPath, []byte(kclModContent), 0o644); err != nil {
			return err
		}

		mainKCLPath := filepath.Join(o.StackDir, util.MainKCLFile)
		mainKCLContent := fmt.Sprint(util.MainKCLFileTemplate)
		if err := os.WriteFile(mainKCLPath, []byte(mainKCLContent), 0o644); err != nil {
			return err
		}
	} else {
		if err := util.CreateWithRefStack(o.Name, o.StackDir, o.CopyFrom); err != nil {
			return err
		}
	}

	fmt.Printf("Created stack '%s' under project directory '%s' successfully\n", o.Name, o.ProjectDir)

	return nil
}

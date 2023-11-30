package build

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	yamlv2 "gopkg.in/yaml.v2"
	"kcl-lang.io/kpm/pkg/api"

	"kusionstack.io/kusion/pkg/apis/project"
	"kusionstack.io/kusion/pkg/apis/stack"
	"kusionstack.io/kusion/pkg/cmd/build/builders"
)

type Options struct {
	IsKclPkg  bool
	Filenames []string
	Flags
}

type Flags struct {
	Output    string
	WorkDir   string
	Arguments map[string]string
	NoStyle   bool
}

const Stdout = "stdout"

func NewBuildOptions() *Options {
	return &Options{
		Filenames: []string{},
		Flags: Flags{
			Arguments: map[string]string{},
		},
	}
}

func (o *Options) Complete(args []string) error {
	o.Filenames = args
	return o.PreSet(stack.IsStack)
}

func (o *Options) Validate() error {
	var wrongFiles []string
	for _, filename := range o.Filenames {
		if filepath.Ext(filename) != ".k" {
			wrongFiles = append(wrongFiles, filename)
		}
	}
	if len(wrongFiles) != 0 {
		return fmt.Errorf("you can only compile files with suffix .k, these are wrong files: %v", wrongFiles)
	}
	return nil
}

func (o *Options) Run() error {
	// Set no style
	if o.NoStyle {
		pterm.DisableStyling()
		pterm.DisableColor()
	}

	// Parse project and stack of work directory
	project, stack, err := project.DetectProjectAndStack(o.WorkDir)
	if err != nil {
		return err
	}

	sp, err := IntentWithSpinner(
		&builders.Options{
			IsKclPkg:  o.IsKclPkg,
			WorkDir:   o.WorkDir,
			Filenames: o.Filenames,
			Arguments: o.Arguments,
			NoStyle:   o.NoStyle,
		},
		project,
		stack,
	)
	if err != nil {
		return err
	}

	yaml, err := yamlv2.Marshal(sp)
	if err != nil {
		return err
	}
	if o.Output == Stdout || o.Output == "" {
		fmt.Print(string(yaml))
	} else {
		if o.WorkDir != "" {
			o.Output = filepath.Join(o.WorkDir, o.Output)
		}

		err = os.WriteFile(o.Output, yaml, 0o666)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Options) PreSet(preCheck func(cur string) bool) error {
	curDir := o.WorkDir
	if o.WorkDir == "" {
		curDir, _ = os.Getwd()
	}
	if ok := preCheck(curDir); !ok {
		if o.Output == "" {
			o.Output = Stdout
		}
		return nil
	}

	if _, err := api.GetKclPackage(o.WorkDir); err == nil {
		o.IsKclPkg = true
		return nil
	}
	return nil
}

package compile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	yamlv2 "gopkg.in/yaml.v2"

	"kusionstack.io/kusion/pkg/cmd/spec"
	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/projectstack"
)

type CompileOptions struct {
	IsCheck   bool
	Filenames []string
	CompileFlags
}

type CompileFlags struct {
	Output      string
	WorkDir     string
	Settings    []string
	Arguments   []string
	Overrides   []string
	DisableNone bool
	OverrideAST bool
	NoStyle     bool
}

const Stdout = "stdout"

func NewCompileOptions() *CompileOptions {
	return &CompileOptions{
		Filenames: []string{},
		CompileFlags: CompileFlags{
			Settings:  []string{},
			Arguments: []string{},
			Overrides: []string{},
		},
	}
}

func (o *CompileOptions) Complete(args []string) {
	o.Filenames = args
	o.PreSet(projectstack.IsStack)
}

func (o *CompileOptions) Validate() error {
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

func (o *CompileOptions) Run() error {
	// Set no style
	if o.NoStyle {
		pterm.DisableStyling()
		pterm.DisableColor()
	}

	// Parse project and stack of work directory
	project, stack, err := projectstack.DetectProjectAndStack(o.WorkDir)
	if err != nil {
		return err
	}

	sp, err := spec.GenerateSpecWithSpinner(&generator.Options{
		WorkDir:     o.WorkDir,
		Filenames:   o.Filenames,
		Settings:    o.Settings,
		Arguments:   o.Arguments,
		Overrides:   o.Overrides,
		DisableNone: o.DisableNone,
		OverrideAST: o.OverrideAST,
		NoStyle:     o.NoStyle,
	}, project, stack)
	if err != nil {
		// only print err in the check command
		if o.IsCheck {
			fmt.Println(err)
			return nil
		} else {
			return err
		}
	}

	yaml, err := yamlv2.Marshal(sp.Resources)
	if err != nil {
		return err
	}
	if o.Output == Stdout {
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

func (o *CompileOptions) PreSet(preCheck func(cur string) bool) {
	curDir := o.WorkDir
	if o.WorkDir == "" {
		curDir, _ = os.Getwd()
	}
	if ok := preCheck(curDir); !ok {
		if o.Output == "" {
			o.Output = Stdout
		}
		return
	}

	if len(o.Settings) == 0 {
		o.Settings = []string{filepath.Join(projectstack.CiTestDir, projectstack.SettingsFile), projectstack.KclFile}
	}

	if o.Output == "" {
		o.Output = filepath.Join(projectstack.CiTestDir, projectstack.StdoutGoldenFile)
	}
}

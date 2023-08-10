package compile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	yamlv2 "gopkg.in/yaml.v2"

	"kusionstack.io/kusion/pkg/cmd/spec"
	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/log"
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

func (o *CompileOptions) Complete(args []string) error {
	o.Filenames = args
	return o.PreSet(projectstack.IsStack)
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

	yaml, err := yamlv2.Marshal(sp)
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

func (o *CompileOptions) PreSet(preCheck func(cur string) bool) error {
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

	if len(o.Settings) == 0 {
		o.Settings = []string{projectstack.KclFile}
		info, err := os.Stat(filepath.Join(curDir, projectstack.CiTestDir, projectstack.SettingsFile))
		switch {
		case err != nil && os.IsNotExist(err):
			log.Warnf("%s is not exist", projectstack.SettingsFile)
		case err != nil && !os.IsNotExist(err):
			return err
		case err == nil && info.Mode().IsRegular():
			o.Settings = append(o.Settings, filepath.Join(projectstack.CiTestDir, projectstack.SettingsFile))
		case err == nil && !info.Mode().IsRegular():
			log.Warnf("%s is not a regular file", projectstack.SettingsFile)
		}
	}

	if o.Output == "" {
		absCiTestDir := filepath.Join(curDir, projectstack.CiTestDir)
		_, err := os.Stat(absCiTestDir)
		if err != nil && os.IsNotExist(err) {
			_ = os.Mkdir(absCiTestDir, 0o750)
		}
		o.Output = filepath.Join(projectstack.CiTestDir, projectstack.StdoutGoldenFile)
	}
	return nil
}

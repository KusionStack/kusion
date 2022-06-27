package compile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"kusionstack.io/kusion/pkg/compile"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/projectstack"
)

type CompileOptions struct {
	IsCheck     bool
	Filenames   []string
	Arguments   []string
	Settings    []string
	Output      string
	WorkDir     string
	DisableNone bool
	OverrideAST bool
	Overrides   []string
	LogToStderr bool
}

func NewCompileOptions() *CompileOptions {
	return &CompileOptions{
		Arguments: []string{},
		Settings:  []string{},
	}
}

func (o *CompileOptions) Complete(args []string) {
	o.Filenames = args
	o.PreSet(projectstack.IsStack)
	if o.LogToStderr {
		log.SetOutput(os.Stderr)
	}
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
	// Compile
	compileResult, err := compile.Compile(o.WorkDir, o.Filenames, o.Settings, o.Arguments, o.Overrides, o.DisableNone, o.OverrideAST)
	if o.IsCheck {
		if err != nil {
			fmt.Print(err)
		}
	} else {
		if err != nil {
			return err
		}

		// Output
		yaml := compileResult.RawYAML()
		if o.Output == "" {
			fmt.Print(yaml)
		} else {
			if o.WorkDir != "" {
				o.Output = filepath.Join(o.WorkDir, o.Output)
			}

			err := ioutil.WriteFile(o.Output, []byte(yaml), 0o666)
			if err != nil {
				return err
			}
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
		return
	}

	if len(o.Settings) == 0 && o.Output == "" && len(o.Filenames) == 0 {
		o.Settings = []string{filepath.Join(projectstack.CiTestDir, projectstack.SettingsFile), projectstack.KclFile}
		o.Output = filepath.Join(projectstack.CiTestDir, projectstack.StdoutGoldenFile)
	}
}

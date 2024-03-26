package build

import (
	"fmt"
	"os"
	"path/filepath"

	"kcl-lang.io/kpm/pkg/api"

	"kusionstack.io/kusion/pkg/project"
)

const (
	KclFile = "kcl.yaml"
)

type Options struct {
	KclPkg    *api.KclPackage
	Filenames []string
	Flags
}

type Flags struct {
	Output    string
	WorkDir   string
	Settings  []string
	Arguments map[string]string
	NoStyle   bool
	Backend   string
	Workspace string
}

const Stdout = "stdout"

func NewBuildOptions() *Options {
	return &Options{
		Filenames: []string{},
		Flags: Flags{
			Arguments: map[string]string{},
			Settings:  make([]string, 0),
		},
	}
}

func (o *Options) Complete(args []string) error {
	o.Filenames = args
	return o.PreSet(project.IsStack)
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

	var err error
	o.KclPkg, err = api.GetKclPackage(o.WorkDir)
	if err != nil {
		return err
	}

	if len(o.Settings) == 0 {
		// if kcl.yaml exists, use it as settings
		if _, err := os.Stat(filepath.Join(o.WorkDir, KclFile)); err == nil {
			o.Settings = []string{KclFile}
		}
	}
	return nil
}

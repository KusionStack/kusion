package init

import (
	"errors"
	"fmt"
	"path/filepath"

	"kusionstack.io/kusion/pkg/cmd/init/util"
	"kusionstack.io/kusion/pkg/scaffold"
)

var ErrNotEmptyArgs = errors.New("no args accepted")

type Options struct {
	Name string
	Flags
}

type Flags struct {
	ProjectDir string
}

func NewOptions() *Options {
	return &Options{
		Flags: Flags{
			ProjectDir: "",
		},
	}
}

func (o *Options) Complete(args []string) error {
	if len(args) > 0 {
		return ErrNotEmptyArgs
	}

	dir, name, err := util.GetDirAndName()
	if err != nil {
		return err
	}

	o.Name = name
	if o.ProjectDir == "" {
		o.ProjectDir = dir
	} else {
		// Use the absolute path of the target project directory.
		absTargetDir, err := filepath.Abs(o.ProjectDir)
		if err != nil {
			return err
		}
		o.ProjectDir = absTargetDir
		o.Name = filepath.Base(o.ProjectDir)
	}

	return nil
}

func (o *Options) Validate() error {
	if err := util.ValidateProjectDir(o.ProjectDir); err != nil {
		return err
	}

	if err := util.ValidateProjectName(o.Name); err != nil {
		return err
	}

	return nil
}

func (o *Options) Run() error {
	if err := scaffold.GenDemoProject(o.ProjectDir, o.Name); err != nil {
		return err
	}

	fmt.Printf("Initiated demo project '%s' successfully\n", o.Name)

	return nil
}

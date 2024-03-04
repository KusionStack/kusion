package init

import (
	"errors"
	"fmt"

	"kusionstack.io/kusion/pkg/cmd/init/util"
	"kusionstack.io/kusion/pkg/scaffold"
)

var ErrNotEmptyArgs = errors.New("no args accepted")

type Options struct {
	Name       string
	ProjectDir string
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) Complete(args []string) error {
	if len(args) > 0 {
		return ErrNotEmptyArgs
	}

	dir, name, err := util.GetDirAndName()
	if err != nil {
		return err
	}

	o.ProjectDir = dir
	o.Name = name

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

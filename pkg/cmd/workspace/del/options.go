package del

import (
	"fmt"

	"kusionstack.io/kusion/pkg/cmd/workspace/util"
	"kusionstack.io/kusion/pkg/workspace"
)

type Options struct {
	Name string
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) Complete(args []string) error {
	name, err := util.GetNameFromArgs(args)
	if err != nil {
		return err
	}
	o.Name = name
	return nil
}

func (o *Options) Validate() error {
	if err := util.ValidateName(o.Name); err != nil {
		return err
	}
	return nil
}

func (o *Options) Run() error {
	if err := workspace.DeleteWorkspaceByDefaultOperator(o.Name); err != nil {
		return err
	}
	fmt.Printf("delete workspace %s successfully\n", o.Name)
	return nil
}

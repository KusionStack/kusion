package show

import (
	"fmt"

	"gopkg.in/yaml.v3"

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
	ws, err := workspace.GetWorkspaceByDefaultOperator(o.Name)
	if err != nil {
		return err
	}
	content, err := yaml.Marshal(ws)
	if err != nil {
		return fmt.Errorf("yaml marshal workspace configuration failed: %w", err)
	}
	fmt.Print(string(content))
	return nil
}

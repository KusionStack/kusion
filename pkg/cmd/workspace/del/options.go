package del

import (
	"fmt"

	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/cmd/workspace/util"
)

type Options struct {
	Name    string
	Backend string
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

func (o *Options) Run() error {
	storage, err := backend.NewWorkspaceStorage(o.Backend)
	if err != nil {
		return err
	}

	// get current workspace name if not specified.
	if o.Name == "" {
		var name string
		name, err = storage.GetCurrent()
		if err != nil {
			return err
		}
		o.Name = name
	}
	if err = util.ValidateNotDefaultName(o.Name); err != nil {
		return err
	}

	if err = storage.Delete(o.Name); err != nil {
		return err
	}
	fmt.Printf("delete workspace %s successfully\n", o.Name)
	return nil
}

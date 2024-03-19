package create

import (
	"fmt"

	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/cmd/workspace/util"
)

type Options struct {
	Name     string
	FilePath string
	Backend  string
	Current  bool
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
	if err := util.ValidateNotDefaultName(o.Name); err != nil {
		return err
	}
	if err := util.ValidateFilePath(o.FilePath); err != nil {
		return err
	}
	return nil
}

func (o *Options) Run() error {
	storage, err := backend.NewWorkspaceStorage(o.Backend)
	if err != nil {
		return err
	}

	ws, err := util.GetValidWorkspaceFromFile(o.FilePath, o.Name)
	if err != nil {
		return err
	}
	if err = storage.Create(ws); err != nil {
		return err
	}

	if o.Current {
		if err = storage.SetCurrent(o.Name); err != nil {
			return err
		}
	}
	fmt.Printf("create workspace %s successfully\n", o.Name)
	return nil
}

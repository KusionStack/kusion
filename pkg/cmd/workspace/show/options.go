package show

import (
	"fmt"

	"gopkg.in/yaml.v3"

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

	ws, err := storage.Get(o.Name)
	if err != nil {
		return err
	}
	content, err := yaml.Marshal(ws)
	if err != nil {
		return fmt.Errorf("yaml marshal workspace configuration failed: %w", err)
	}
	fmt.Printf("show configuration of workspace %s:\n", ws.Name)
	fmt.Print(string(content))
	return nil
}

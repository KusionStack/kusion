package list

import (
	"fmt"

	"kusionstack.io/kusion/pkg/backend"
)

type Options struct {
	Backend string
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) Validate(args []string) error {
	if len(args) != 0 {
		return ErrNotEmptyArgs
	}
	return nil
}

func (o *Options) Run() error {
	storage, err := backend.NewWorkspaceStorage(o.Backend)
	if err != nil {
		return err
	}

	names, err := storage.GetNames()
	if err != nil {
		return err
	}
	current, err := storage.GetCurrent()
	if err != nil {
		return err
	}

	for _, name := range names {
		if name == current {
			fmt.Println("* " + name)
		} else {
			fmt.Println("- " + name)
		}
	}
	return nil
}

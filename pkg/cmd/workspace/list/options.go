package list

import (
	"fmt"

	"gopkg.in/yaml.v3"

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
	content, err := yaml.Marshal(names)
	if err != nil {
		return fmt.Errorf("yaml marshal workspace names failed: %w", err)
	}
	fmt.Print(string(content))
	return nil
}

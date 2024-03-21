package unset

import (
	"fmt"

	"kusionstack.io/kusion/pkg/cmd/config/util"
	"kusionstack.io/kusion/pkg/config"
)

type Options struct {
	Item string
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) Complete(args []string) error {
	item, err := util.GetItemFromArgs(args)
	if err != nil {
		return err
	}
	o.Item = item
	return nil
}

func (o *Options) Validate() error {
	if err := util.ValidateItem(o.Item); err != nil {
		return err
	}
	return nil
}

func (o *Options) Run() error {
	if err := config.DeleteConfigItem(o.Item); err != nil {
		return err
	}

	fmt.Printf("set config item %s successfully", o.Item)
	return nil
}

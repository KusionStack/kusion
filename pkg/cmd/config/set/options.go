package set

import (
	"fmt"

	"kusionstack.io/kusion/pkg/cmd/config/util"
	"kusionstack.io/kusion/pkg/config"
)

type Options struct {
	Item  string
	Value string
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) Complete(args []string) error {
	item, value, err := util.GetItemValueFromArgs(args)
	if err != nil {
		return err
	}
	o.Item = item
	o.Value = value
	return nil
}

func (o *Options) Validate() error {
	if err := util.ValidateItem(o.Item); err != nil {
		return err
	}
	if err := util.ValidateValue(o.Value); err != nil {
		return err
	}
	return nil
}

func (o *Options) Run() error {
	if err := config.SetEncodedConfigItem(o.Item, o.Value); err != nil {
		return err
	}

	fmt.Printf("set config item %s successfully", o.Item)
	return nil
}

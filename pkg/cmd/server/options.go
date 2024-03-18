package server

import "kusionstack.io/kusion/pkg/server/route"

type Options struct {
	Mode string
}

func NewServerOptions() *Options {
	return &Options{
		Mode: "KCP",
	}
}

func (o *Options) Complete(args []string) {}

func (o *Options) Validate() error {
	return nil
}

func (o *Options) Run() error {
	if _, err := route.NewCoreRoute(); err == nil {
		return nil
	}
	return nil
}

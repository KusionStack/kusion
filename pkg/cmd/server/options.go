package server

import (
	"kusionstack.io/kusion/pkg/server"
	"kusionstack.io/kusion/pkg/server/route"
)

func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		Mode:     "KCP",
		Database: DatabaseOptions{},
	}
}

func (o *ServerOptions) Complete(args []string) {}

func (o *ServerOptions) Validate() error {
	return nil
}

func (o *ServerOptions) Config() (*server.Config, error) {
	cfg := server.NewConfig()
	o.Database.ApplyTo(cfg)
	return cfg, nil
}

func (o *ServerOptions) Run() error {
	config, err := o.Config()
	if err != nil {
		return err
	}
	if _, err := route.NewCoreRoute(config); err == nil {
		return nil
	}
	return nil
}

package server

import (
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/server"
	"kusionstack.io/kusion/pkg/server/route"
)

func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		Mode:               DefaultMode,
		Port:               DefaultPort,
		AuthEnabled:        false,
		AuthWhitelist:      []string{},
		AuthKeyType:        DefaultAuthKeyType,
		Database:           DatabaseOptions{},
		DefaultBackend:     DefaultBackendOptions{},
		DefaultSource:      DefaultSourceOptions{},
		MaxConcurrent:      constant.MaxConcurrent,
		MaxAsyncConcurrent: constant.MaxAsyncConcurrent,
		MaxAsyncBuffer:     constant.MaxAsyncBuffer,
		LogFilePath:        constant.DefaultLogFilePath,
	}
}

func (o *ServerOptions) Complete(args []string) {}

func (o *ServerOptions) Validate() error {
	return nil
}

func (o *ServerOptions) Config() (*server.Config, error) {
	cfg := server.NewConfig()
	o.Database.ApplyTo(cfg)
	o.DefaultBackend.ApplyTo(cfg)
	o.DefaultSource.ApplyTo(cfg)
	cfg.Port = o.Port
	cfg.AuthEnabled = o.AuthEnabled
	cfg.AuthWhitelist = o.AuthWhitelist
	cfg.AuthKeyType = o.AuthKeyType
	cfg.MaxConcurrent = o.MaxConcurrent
	cfg.MaxAsyncConcurrent = o.MaxAsyncConcurrent
	cfg.MaxAsyncBuffer = o.MaxAsyncBuffer
	cfg.LogFilePath = o.LogFilePath
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

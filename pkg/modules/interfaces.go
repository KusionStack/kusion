package modules

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules/proto"
)

const PluginKey = "module-default"

// Generator is an interface for things that can generate Intent from input configurations.
// todo it's for built-in generators and we should consider to convert it to a general Module interface
type Generator interface {
	// Generate performs the intent generate operation.
	Generate(intent *v1.Intent) error
}

// Module is the interface that we're exposing as a kusion module plugin.
type Module interface {
	Generate(req *proto.GeneratorRequest) (*proto.GeneratorResponse, error)
}

// NewGeneratorFunc is a function that returns a Generator.
type NewGeneratorFunc func() (Generator, error)

type GRPCClient struct {
	client proto.ModuleClient
}

func (c *GRPCClient) Generate(req *proto.GeneratorRequest) (*proto.GeneratorResponse, error) {
	return c.client.Generate(context.Background(), req)
}

type GRPCServer struct {
	// This is the real implementation
	Impl Module
	proto.UnimplementedModuleServer
}

func (s *GRPCServer) Generate(ctx context.Context, req *proto.GeneratorRequest) (res *proto.GeneratorResponse, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.WithStack(err)
			res = &proto.GeneratorResponse{}
		}
	}()
	res, err = s.Impl.Generate(req)
	return
}

type GRPCPlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins that are written in Go.
	Impl Module
}

// GRPCServer is going to be invoked by the go-plugin framework
func (p *GRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterModuleServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

// GRPCClient is going to be invoked by the go-plugin framework
func (p *GRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewModuleClient(c)}, nil
}

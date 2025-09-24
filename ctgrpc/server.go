package ctgrpc

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/surkovvs/ct/ctifaces"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	Network string
	Address string
	server  *grpc.Server
}

func NewServer(cfg ctifaces.GRPCServerConfigurator, opt ...grpc.ServerOption) Server {
	var credsOpt grpc.ServerOption
	if tls := cfg.GetTLS(); tls != nil {
		credsOpt = grpc.Creds(credentials.NewTLS(tls))
	} else {
		credsOpt = grpc.Creds(insecure.NewCredentials())
	}
	applyOpts := []grpc.ServerOption{credsOpt}
	if len(opt) == 0 {
		applyOpts = append(applyOpts, defaultGRPCServerOpts()...)
	} else {
		applyOpts = append(applyOpts, opt...)
	}

	return Server{
		Network: cfg.GetNetwork(),
		Address: cfg.GetAddress(),
		server:  grpc.NewServer(applyOpts...),
	}
}

func (s Server) GetServer() *grpc.Server {
	return s.server
}

func (s Server) RegisterService(sd *grpc.ServiceDesc, ss any) {
	s.server.RegisterService(sd, ss)
}

func (s Server) Run(_ context.Context) error {
	l, err := net.Listen(s.Network, s.Address)
	if err != nil {
		return fmt.Errorf("net listen: %w", err)
	}

	err = s.server.Serve(l)
	if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return fmt.Errorf("grpc serve: %w", err)
	}
	return nil
}

func (s Server) Shutdown(ctx context.Context) error {
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		s.server.GracefulStop()
	}()
	select {
	case <-ctx.Done():
		s.server.Stop()
		return errors.New(
			"grpc server graceful stop exeeded deadline, stop is called")
	case <-stopped:
		return nil
	}
}

func (s Server) GetModuleNamePrefix() string {
	return "grpc_server"
}

package cthttp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/surkovvs/ct/ctifaces"
)

const defaultReadHeaderTimeout = time.Second * 5

type Server struct {
	Network string
	Address string
	server  *http.Server
}

func NewServer(cfg ctifaces.HTTPServerConfigurator) Server {
	return Server{
		server: &http.Server{
			ReadHeaderTimeout: defaultReadHeaderTimeout,
			TLSConfig:         cfg.GetTLS(),
		},
		Network: cfg.GetNetwork(),
		Address: cfg.GetAddress(),
	}
}

func (s Server) GetServer() *http.Server {
	return s.server
}

func (s Server) SetHandler(handler http.Handler) {
	s.server.Handler = handler
}

func (s Server) Run(_ context.Context) error {
	l, err := net.Listen(s.Network, s.Address)
	if err != nil {
		return fmt.Errorf("net listen: %w", err)
	}

	err = s.server.Serve(l)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http serve: %w", err)
	}
	return nil
}

func (s Server) Shutdown(ctx context.Context) error {
	err := s.server.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("http server shutdown: %w", err)
	}
	return nil
}

func (_ Server) GetModuleNamePrefix() string {
	return "http_server"
}

func (_ Server) PreidentifyModuleGroup() string {
	return "ingress"
}

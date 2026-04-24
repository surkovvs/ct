package ctgrpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/surkovvs/ct/ctifaces"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientProvider interface {
	GetClient() *grpc.ClientConn
}

var _ ClientProvider = (*Client)(nil)

type Client struct {
	Address string
	opts    []grpc.DialOption
	client  *grpc.ClientConn
}

func NewClient(cfg ctifaces.GRPCClientConfigurator, opt ...grpc.DialOption) *Client {
	var credsOpt grpc.DialOption
	if tls := cfg.GetTLS(); tls != nil {
		credsOpt = grpc.WithTransportCredentials(credentials.NewTLS(tls))
	} else {
		credsOpt = grpc.WithTransportCredentials(insecure.NewCredentials())
	}
	applyOpts := []grpc.DialOption{credsOpt}
	if len(opt) == 0 {
		applyOpts = append(applyOpts, defaultGRPCClientOpts()...)
	} else {
		applyOpts = append(applyOpts, defaultGRPCClientOpts()...)
	}

	return &Client{
		Address: cfg.GetAddress(),
		opts:    applyOpts,
		client:  nil,
	}
}

func (c *Client) GetClient() *grpc.ClientConn {
	return c.client
}

func (c *Client) Init(ctx context.Context) error {
	var err error
	c.client, err = grpc.NewClient(c.Address, c.opts...)
	if err != nil {
		return fmt.Errorf("new grpc client: %w", err)
	}

	for {
		state := c.client.GetState()
		if state == connectivity.Ready || state == connectivity.Idle {
			return nil
		}
		if !c.client.WaitForStateChange(ctx, state) {
			return errors.New("timeout waiting for grpc conn ready, current status: " + c.client.GetState().String())
		}
	}
}

func (c *Client) Shutdown(_ context.Context) error {
	err := c.client.Close()
	if err != nil {
		return fmt.Errorf("close grpc client: %w", err)
	}
	return nil
}

func (_ *Client) GetModuleNamePrefix() string {
	return "grpc_client"
}

func (_ *Client) PreidentifyModuleGroup() string {
	return "egress"
}

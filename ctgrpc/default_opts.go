package ctgrpc

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	defaultConnectionTimeout = 10 * time.Second
	defaultMaxRecvMsgSize    = 4 << 20 // 4 MB
	defaultMaxSendMsgSize    = 4 << 20 // 4 MB

	defaultKeepaliveServerPingInterval    = 30 * time.Second
	defaultKeepaliveServerPingTimeout     = 10 * time.Second
	defaultKeepaliveMaxConnectionIdle     = 5 * time.Minute
	defaultKeepaliveMaxConnectionAge      = 15 * time.Minute
	defaultKeepaliveMaxConnectionAgeGrace = 5 * time.Minute

	defaultKeepaliveMinTimePingIntervalServerSide = 10 * time.Second
	defaultKeepalivePermWithoutStream             = true

	defaultKeepaliveClientPingInterval = 30 * time.Second
	defaultKeepaliveClientPingTimeout  = 10 * time.Second
)

func defaultGRPCServerOpts() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.ConnectionTimeout(defaultConnectionTimeout),
		grpc.MaxRecvMsgSize(defaultMaxRecvMsgSize),
		grpc.MaxSendMsgSize(defaultMaxSendMsgSize),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:                  defaultKeepaliveServerPingInterval,
			Timeout:               defaultKeepaliveServerPingTimeout,
			MaxConnectionIdle:     defaultKeepaliveMaxConnectionIdle,
			MaxConnectionAge:      defaultKeepaliveMaxConnectionAge,
			MaxConnectionAgeGrace: defaultKeepaliveMaxConnectionAgeGrace,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             defaultKeepaliveMinTimePingIntervalServerSide,
			PermitWithoutStream: defaultKeepalivePermWithoutStream,
		}),
	}
}

func defaultGRPCClientOpts() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                defaultKeepaliveClientPingInterval,
			Timeout:             defaultKeepaliveClientPingTimeout,
			PermitWithoutStream: defaultKeepalivePermWithoutStream,
		}),
	}
}

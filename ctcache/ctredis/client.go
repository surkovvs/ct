package ctredis

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/surkovvs/ct/ctifaces"
)

type Client struct {
	uc redis.UniversalClient
}

func NewClient(cfg ctifaces.RedisClientConfigurator) *Client {
	return &Client{
		uc: redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:      cfg.GetAddresses(),
			ClientName: cfg.GetClientName(),
			DB:         cfg.GetDB(),
			Username:   cfg.GetUsername(),
			Password:   cfg.GetPassword(),
			TLSConfig:  cfg.GetTLS(),
		}),
	}
}

func (c Client) Init(ctx context.Context) error {
	return c.uc.Ping(ctx).Err()
}

func (c Client) Shutdown(ctx context.Context) error {
	return c.uc.Close()
}

func (c Client) GetRedisClient() redis.UniversalClient {
	return c.uc
}

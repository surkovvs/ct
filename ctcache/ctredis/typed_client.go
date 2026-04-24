package ctredis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Codec[K, V any] interface {
	Key(key K) (string, error)
	Encode(val V) (string, error)
	Decode(val string) (V, error)
}

type TypedClient[K, V any] struct {
	uc    redis.UniversalClient
	codec Codec[K, V]
}

func NewTypedClient[K, V any](c Client, codec Codec[K, V]) TypedClient[K, V] {
	return TypedClient[K, V]{
		uc:    c.uc,
		codec: codec,
	}
}

func (c TypedClient[K, V]) Set(ctx context.Context, key K, val V, exp time.Duration) error {
	k, err := c.codec.Key(key)
	if err != nil {
		return fmt.Errorf("later: %w", err)
	}
	v, err := c.codec.Encode(val)
	if err != nil {
		return fmt.Errorf("later: %w", err)
	}
	return c.uc.Set(ctx, k, v, exp).Err()
}

func (c TypedClient[K, V]) Get(ctx context.Context, key K) (V, error) {
	var zero V
	k, err := c.codec.Key(key)
	if err != nil {
		return zero, fmt.Errorf("later: %w", err)
	}
	raw, err := c.uc.Get(ctx, k).Result()
	if err != nil {
		return zero, fmt.Errorf("redis get: %w", err)
	}
	res, err := c.codec.Decode(raw)
	if err != nil {
		return zero, fmt.Errorf("later: %w", err)
	}
	return res, nil
}

func (c TypedClient[K, V]) Del(ctx context.Context, keys ...K) error {
	ks := make([]string, 0, len(keys))
	for _, key := range keys {
		k, err := c.codec.Key(key)
		if err != nil {
			return fmt.Errorf("later: %w", err)
		}
		ks = append(ks, k)
	}
	return c.uc.Del(ctx, ks...).Err()
}

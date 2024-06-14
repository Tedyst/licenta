package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type redisCacheProvider[T any] struct {
	client *redis.Client
	prefix string
}

func NewRedisCacheProvider[T any](client *redis.Client, prefix string) (*redisCacheProvider[T], error) {
	return &redisCacheProvider[T]{
		client: client,
		prefix: prefix,
	}, nil
}

func (r *redisCacheProvider[T]) Get(ctx context.Context, key string) (T, bool, error) {
	var result T
	err := r.client.Get(ctx, r.prefix+key).Scan(result)
	if err != nil {
		return result, false, err
	}

	return result, true, nil
}

func (r *redisCacheProvider[T]) Set(ctx context.Context, key string, value T) error {
	return r.client.Set(ctx, key, value, 0).Err()
}

func (r *redisCacheProvider[T]) Invalidate(ctx context.Context, pattern string) error {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err := r.client.Del(ctx, key).Err(); err != nil {
			return err
		}
	}

	return nil
}

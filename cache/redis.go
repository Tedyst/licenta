package cache

import (
	"context"
	"encoding/json"

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
	var resultText string
	err := r.client.Get(ctx, r.prefix+key).Scan(resultText)
	if err == redis.Nil {
		return result, false, nil
	} else if err != nil {
		return result, false, err
	}

	err = json.Unmarshal([]byte(resultText), &result)
	if err != nil {
		return result, false, err
	}

	return result, true, nil
}

func (r *redisCacheProvider[T]) Set(ctx context.Context, key string, value T) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, valueBytes, 0).Err()
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

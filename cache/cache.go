package cache

import "context"

type CacheProvider[T any] interface {
	Get(ctx context.Context, key string) (T, bool, error)
	Set(ctx context.Context, key string, value T) error
	Invalidate(ctx context.Context, pattern string) error
}

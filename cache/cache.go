package cache

type CacheProvider[T any] interface {
	Get(key string) (T, bool, error)
	Set(key string, value T) error
	Invalidate(key string) error
}

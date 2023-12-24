package cache

import lru "github.com/hashicorp/golang-lru/v2"

const localCacheSize = 1000

type localCacheProvider[T any] struct {
	cache *lru.Cache[string, T]
}

func NewLocalCacheProvider[T any]() (*localCacheProvider[T], error) {
	l, err := lru.New[string, T](localCacheSize)
	if err != nil {
		return nil, err
	}

	return &localCacheProvider[T]{
		cache: l,
	}, nil
}

func (l *localCacheProvider[T]) Get(key string) (T, bool, error) {
	if v, ok := l.cache.Get(key); ok {
		return v, true, nil
	}

	return any(nil).(T), false, nil
}

func (l *localCacheProvider[T]) Set(key string, value T) error {
	l.cache.Add(key, value)
	return nil
}

func (l *localCacheProvider[T]) Invalidate(key string) error {
	l.cache.Remove(key)
	return nil
}

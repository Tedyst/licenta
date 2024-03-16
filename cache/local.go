package cache

import (
	"regexp"

	lru "github.com/hashicorp/golang-lru/v2"
)

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

	var result T
	return result, false, nil
}

func (l *localCacheProvider[T]) Set(key string, value T) error {
	l.cache.Add(key, value)
	return nil
}

func (l *localCacheProvider[T]) Invalidate(pattern string) error {
	regexp, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	for _, k := range l.cache.Keys() {
		if regexp.MatchString(k) {
			l.cache.Remove(k)
		}
	}
	return nil
}

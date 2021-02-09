package internal

import "time"

type Options struct {
	Redis Rediser

	// DefaultTTL is the cache expiration time.
	// 1 hour by default
	DefaultTTL time.Duration

	Marshaller Marshaller

	AbsentKeysLoader func(absentKeys ...string) (interface{}, error)

	ItemToCacheKey func(it interface{}) string

	CacheKeyToMapKey func(cacheKey string) string

	AddCacheMissErrors bool
}

func (opt Options) redisTTL(itemTTL time.Duration) time.Duration {
	if itemTTL < 0 {
		return 0
	}
	if itemTTL < time.Second {
		return opt.DefaultTTL
	}
	return itemTTL
}

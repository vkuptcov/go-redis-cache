package internal

import "time"

type Options struct {
	Redis Rediser

	// DefaultTTL is the cache expiration time.
	// 1 hour by default
	DefaultTTL time.Duration

	Marshaller Marshaller

	AbsentKeysLoader func(absentKeys ...string) (interface{}, error)

	CacheKeyExtractor func(it interface{}) (key, field string)

	TransformCacheKeyForDestination func(key, field string, val interface{}) (newKey, newField string, skip bool)

	AddCacheMissErrors bool

	DisableCacheMissErrorsForSingleElementDst bool
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

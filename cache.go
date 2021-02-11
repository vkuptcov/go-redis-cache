package cache

import (
	"context"
	"time"

	"github.com/vkuptcov/go-redis-cache/v8/internal"
)

type Cache struct {
	opt Options
}

const defaultDuration = 1 * time.Hour

func NewCache(opt Options) *Cache {
	cacheDuration := defaultDuration
	if opt.DefaultTTL >= 1*time.Second {
		cacheDuration = opt.DefaultTTL
	}
	opt.DefaultTTL = cacheDuration

	return &Cache{
		opt: opt,
	}
}

func (cd *Cache) WithTTL(ttl time.Duration) *Cache {
	opts := cd.opt
	opts.DefaultTTL = ttl
	return &Cache{opt: opts}
}

func (cd *Cache) WithAbsentKeysLoader(f func(absentKeys ...string) (interface{}, error)) *Cache {
	opts := cd.opt
	opts.AbsentKeysLoader = f
	return &Cache{opt: opts}
}

func (cd *Cache) WithItemToCacheKey(f func(it interface{}) (key, field string)) *Cache {
	opts := cd.opt
	opts.ItemToCacheKey = f
	return &Cache{opt: opts}
}

func (cd *Cache) ConvertCacheKeyToMapKey(f func(cacheKey string) string) *Cache {
	opts := cd.opt
	opts.CacheKeyToMapKey = f
	return &Cache{opt: opts}
}

func (cd *Cache) AddCacheMissErrors() *Cache {
	opts := cd.opt
	opts.AddCacheMissErrors = true
	return &Cache{opt: opts}
}

// Set sets multiple elements
func (cd *Cache) Set(ctx context.Context, items ...*Item) error {
	return internal.SetMulti(ctx, cd.opt, items...)
}

func (cd *Cache) SetKV(ctx context.Context, keyValPairs ...interface{}) error {
	return internal.SetKV(ctx, cd.opt, keyValPairs...)
}

func (cd *Cache) HSetKV(ctx context.Context, key string, fieldValPairs ...interface{}) error {
	return internal.HSetKV(ctx, cd.opt, key, fieldValPairs...)
}

// Get gets the value for the given keys
func (cd *Cache) Get(ctx context.Context, dst interface{}, keys ...string) error {
	return internal.Get(ctx, cd.opt, dst, keys)
}

func (cd *Cache) HGetAll(ctx context.Context, dst interface{}, keys ...string) error {
	return internal.HGetAll(ctx, cd.opt, dst, keys)
}

func (cd *Cache) HGetFieldsForKey(ctx context.Context, dst interface{}, key string, fields ...string) error {
	return internal.HGetFields(ctx, cd.opt, dst, map[string][]string{key: fields})
}

func (cd *Cache) HGetKeysAndFields(ctx context.Context, dst interface{}, keysToFields map[string][]string) error {
	return internal.HGetFields(ctx, cd.opt, dst, keysToFields)
}

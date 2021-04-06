package cache

import (
	"context"
	"time"

	"github.com/vkuptcov/go-redis-cache/v8/internal"
)

type Cache struct {
	opt Options
}

const DefaultDuration = 1 * time.Hour

func NewCache(opt Options) *Cache {
	cacheDuration := DefaultDuration
	if opt.DefaultTTL >= 1*time.Second {
		cacheDuration = opt.DefaultTTL
	}
	opt.DefaultTTL = cacheDuration

	return &Cache{
		opt: opt,
	}
}

// WithTTL overrides the TTL which is set on Cache creation via cache.Options
// If it is in the [0, 1s) than the value from cache.Options will be used
// If it is less than 0, than cached values will be stored without an explicit TTL
func (cd *Cache) WithTTL(ttl time.Duration) *Cache {
	opts := cd.opt
	opts.DefaultTTL = ttl
	return &Cache{opt: opts}
}

// WithAbsentKeysLoader sets a function to load absent keysToLoad.
// It returns a slice or a map of string keysToLoad to an item to cache.
// The returned item might be something, which can be cached by a codec or
// an instance of Item
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

func (cd *Cache) TransformCacheKeyForDestination(f func(key, field string, val interface{}) (newKey, newField string, skip bool)) *Cache {
	opts := cd.opt
	opts.TransformCacheKeyForDestination = f
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

// Get gets the value for the given keysToLoad
// dst might be
// 1. single element such as structure/number/string/etc
// 2. a slice of single elements
// 3. a map key to a single element
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

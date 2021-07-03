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

// ExtractCacheKeyWith sets a function which is used to transform loaded item
// into key and an optional field.
// It must be set if AbsentKeysLoader function returns a slice of elements to be cached.
// Otherwise it's not possible to determine how to cache the returned elements.
func (cd *Cache) ExtractCacheKeyWith(f func(it interface{}) (key, field string)) *Cache {
	opts := cd.opt
	opts.CacheKeyExtractor = f
	return &Cache{opt: opts}
}

// TransformCacheKeyForDestination changes the data which is used to create a key for a destination map.
// By default it's created from the same keys which are used as cache keys.
// If returned skip parameter is true then the returned element is cached but isn't added into destination.
func (cd *Cache) TransformCacheKeyForDestination(f func(key, field string, val interface{}) (newKey, newField string, skip bool)) *Cache {
	opts := cd.opt
	opts.TransformCacheKeyForDestination = f
	return &Cache{opt: opts}
}

// AddCacheMissErrors makes all *Get methods include ErrCacheMiss errors into returned *KeyErr
// for keys which aren't found in cache.
// By default they aren't included in case we load something in a slice or a map
func (cd *Cache) AddCacheMissErrors() *Cache {
	opts := cd.opt
	opts.AddCacheMissErrors = true
	return &Cache{opt: opts}
}

// DisableCacheMissErrorsForSingleElementDst suppresses returning ErrCacheMiss
// if the desired cache key isn't found and destination is NOT a slice or a map
func (cd *Cache) DisableCacheMissErrorsForSingleElementDst() *Cache {
	opts := cd.opt
	opts.DisableCacheMissErrorsForSingleElementDst = true
	return &Cache{opt: opts}
}

// Set sets multiple items in cache.
// As the entire Item needs to be specified,
// it's possible to mix different types and keys, use hash maps, set custom TTL and so on
func (cd *Cache) Set(ctx context.Context, items ...*Item) error {
	return internal.SetMulti(ctx, cd.opt, items...)
}

// SetKV sets multiple items in cache, default TTL or the TTL from WithTTL will be used
func (cd *Cache) SetKV(ctx context.Context, keyValPairs ...interface{}) error {
	return internal.SetKV(ctx, cd.opt, keyValPairs...)
}

// HSetKV sets multiple fields for a single key in a Redis hash map.
// Default TTL or the TTL from WithTTL will be used.
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

// HGetAll loads all fields from Redis hash maps defined for keys
func (cd *Cache) HGetAll(ctx context.Context, dst interface{}, keys ...string) error {
	return internal.HGetAll(ctx, cd.opt, dst, keys)
}

// HGetFieldsForKey loads specified fields from the Redis hash map defined by key
func (cd *Cache) HGetFieldsForKey(ctx context.Context, dst interface{}, key string, fields ...string) error {
	return internal.HGetFields(ctx, cd.opt, dst, map[string][]string{key: fields})
}

// HGetKeysAndFields loads specified fields from the Redis hash map for keys and specified fields
func (cd *Cache) HGetKeysAndFields(ctx context.Context, dst interface{}, keysToFields map[string][]string) error {
	return internal.HGetFields(ctx, cd.opt, dst, keysToFields)
}

func (cd *Cache) Delete(ctx context.Context, keys ...string) error {
	return internal.Delete(ctx, cd.opt, keys)
}

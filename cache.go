package cache

import (
	"context"
	"unsafe"

	"github.com/vkuptcov/go-redis-cache/v8/internal"
)

type Cache struct {
	opt *Options
}

func NewCache(opt *Options) *Cache {
	opt.init()
	return &Cache{
		opt: opt,
	}
}

// Set sets multiple elements
func (cd *Cache) Set(ctx context.Context, items ...*Item) (err error) {
	internalItems := *(*[]*internal.Item)(unsafe.Pointer(&items))
	return internal.SetMulti(ctx, (*internal.Options)(cd.opt), internalItems...)
}

func (cd *Cache) SetKV(ctx context.Context, keyValPairs ...interface{}) (err error) {
	return internal.SetKV(ctx, (*internal.Options)(cd.opt), keyValPairs...)
}

// Get gets the value for the given keys
func (cd *Cache) Get(ctx context.Context, dst interface{}, keys ...string) error {
	return internal.Get(ctx, (*internal.Options)(cd.opt), dst, keys)
}

func (cd *Cache) GetOrLoad(ctx context.Context, dst interface{}, loadFn func(absentKeys ...string) (interface{}, error), keys ...string) error {
	return internal.GetOrLoad(ctx, (*internal.Options)(cd.opt), dst, loadFn, keys...)
}

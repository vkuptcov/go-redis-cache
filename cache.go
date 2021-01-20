package cache

import (
	"context"

	"github.com/vkuptcov/go-redis-cache/v8/internal"
)

type Cache struct {
	opt *Options
}

var _ internal.Marshaller = &Cache{}

func NewCache(opt *Options) *Cache {
	opt.init()
	return &Cache{
		opt: opt,
	}
}

func (cd *Cache) Marshal(value interface{}) ([]byte, error) {
	return cd.opt.Marshaller.Marshal(value)
}

func (cd *Cache) Unmarshal(data []byte, dst interface{}) error {
	return cd.opt.Marshaller.Unmarshal(data, dst)
}

// Set sets multiple elements
func (cd *Cache) Set(ctx context.Context, items ...*Item) (err error) {
	return cd.setMulti(ctx, items...)
}

func (cd *Cache) SetKV(ctx context.Context, keyValPairs ...interface{}) (err error) {
	return cd.setKV(ctx, keyValPairs...)
}

// Get gets the value for the given keys
func (cd *Cache) Get(ctx context.Context, dst interface{}, keys ...string) error {
	return cd.get(ctx, dst, keys)
}

func (cd *Cache) GetOrLoad(ctx context.Context, dst interface{}, loadFn func(absentKeys ...string) (interface{}, error), keys ...string) error {
	return cd.getOrLoad(ctx, dst, loadFn, keys...)
}

package cache

import (
	"context"
	"github.com/pkg/errors"
	"time"

	"github.com/go-redis/redis/v8"
)

type rediser interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	SetXX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd

	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd

	Pipeline() redis.Pipeliner
}

type Options struct {
	Redis rediser

	// DefaultTTL is the cache expiration time.
	// 1 hour by default
	DefaultTTL time.Duration

	Marshaller Marshaller
}

func (opt *Options) init() {
	cacheDuration := time.Hour
	if opt.DefaultTTL >= 1*time.Second {
		cacheDuration = opt.DefaultTTL
	}
	opt.DefaultTTL = cacheDuration
}

type Item struct {
	Key string

	// Value to be cached
	Value interface{}

	// Load returns value to be cached.
	Load func(*Item) (interface{}, error)

	// TTL is the cache expiration time.
	// Default TTL is taken from Options
	TTL time.Duration

	// IfExists only sets the key if it already exist.
	IfExists bool

	// IfNotExists only sets the key if it does not already exist.
	// Only one of IfExists/IfNotExists can be set
	IfNotExists bool
}

func (item *Item) value() (interface{}, error) {
	if item.Load != nil {
		return item.Load(item)
	}
	if item.Value != nil {
		return item.Value, nil
	}
	return nil, nil
}

type Cache struct {
	opt *Options
}

var _ Marshaller = &Cache{}

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
	r := cd.opt.Redis
	var pipeliner redis.Pipeliner
	if len(items) > 1 && r != nil {
		pipeliner = cd.opt.Redis.Pipeline()
		r = pipeliner
	}
	for _, item := range items {
		err = cd.set(ctx, r, item)
		if err != nil {
			return err
		}
	}
	if pipeliner != nil {
		_, err = pipeliner.Exec(ctx)
	}
	return err
}

func (cd *Cache) SetKV(ctx context.Context, keyValPairs ...interface{}) (err error) {
	if len(keyValPairs)%2 != 0 {
		return errors.New("key-values pairs must be provided")
	}
	items := make([]*Item, len(keyValPairs)/2)
	for id := 0; id < len(keyValPairs); id += 2 {
		key, ok := keyValPairs[id].(string)
		if !ok {
			return errors.Errorf("string key expected for position %d, `%#+v` of type %T given", id, keyValPairs[id])
		}
		items[id/2] = &Item{
			Key:   key,
			Value: keyValPairs[id+1],
			TTL:   cd.opt.DefaultTTL,
		}
	}
	return cd.Set(ctx, items...)
}

func (cd *Cache) set(ctx context.Context, redis rediser, item *Item) error {
	value, loadValErr := item.value()
	if loadValErr != nil {
		return loadValErr
	}

	b, marshalErr := cd.Marshal(value)
	if marshalErr != nil {
		return marshalErr
	}

	if item.IfExists {
		return redis.SetXX(ctx, item.Key, b, cd.redisTTL(item)).Err()
	}

	if item.IfNotExists {
		return redis.SetNX(ctx, item.Key, b, cd.redisTTL(item)).Err()
	}

	return redis.Set(ctx, item.Key, b, cd.redisTTL(item)).Err()
}

func (cd *Cache) redisTTL(item *Item) time.Duration {
	if item.TTL < 0 {
		return 0
	}
	if item.TTL < time.Second {
		return cd.opt.DefaultTTL
	}
	return item.TTL
}

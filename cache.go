package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/go-redis/redis/v8"
)

var ErrCacheMiss = errors.New("cache: key is missing")

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
	if item.Value != nil {
		return item.Value, nil
	}
	if item.Load != nil {
		val, err := item.Load(item)
		item.Value = val
		return val, err
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
			return errors.Errorf("string key expected for position %d, `%#+v` of type %T given", id, keyValPairs[id], keyValPairs[id])
		}
		items[id/2] = &Item{
			Key:   key,
			Value: keyValPairs[id+1],
			TTL:   cd.opt.DefaultTTL,
		}
	}
	return cd.Set(ctx, items...)
}

// Get gets the value for the given key.
func (cd *Cache) Get(ctx context.Context, dst interface{}, keys ...string) error {
	return cd.get(ctx, dst, keys)
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

func (cd *Cache) get(ctx context.Context, dst interface{}, keys []string) error {
	loadedBytes, loadedElementsCount, loadErr := cd.getBytes(ctx, keys)

	if len(keys) == 1 {
		if loadErr != nil {
			return loadErr
		}
		return cd.Unmarshal(loadedBytes[0], dst)
	}
	container, containerErr := newContainer(dst)
	if containerErr != nil {
		return containerErr
	}
	container.initWithSize(loadedElementsCount)
	for idx, b := range loadedBytes {
		dstEl := container.dstEl()
		unmarshalErr := cd.Unmarshal(b, dstEl)
		if unmarshalErr != nil {
			// @todo init and add KeyErr
			return unmarshalErr
		}
		container.addElement(keys[idx], dstEl)
	}
	return nil
}

// @todo
// 1. optimize it for single key
func (cd *Cache) getBytes(ctx context.Context, keys []string) (b [][]byte, loadedElementsCount int, err error) {
	includeCacheMissErrors, _ := ctx.Value(includeCacheMissErrsKey).(bool)
	pipeliner := cd.opt.Redis.Pipeline()
	for _, k := range keys {
		_ = pipeliner.Get(ctx, k)
	}

	// errors are handled by keys
	cmds, _ := pipeliner.Exec(ctx)

	b = make([][]byte, len(keys))

	keysToErrs := map[string]error{}

	for idx, cmd := range cmds {
		k := keys[idx]
		var keyErr error
		switch {
		case cmd.Err() == nil:
			if strCmd, ok := cmd.(*redis.StringCmd); ok {
				b[idx], keyErr = strCmd.Bytes()
			} else {
				keyErr = errors.Errorf("*redis.StringCmd expected for key `%s`, %T received", k, cmd)
			}
		case errors.Is(cmd.Err(), redis.Nil):
			if includeCacheMissErrors {
				keyErr = ErrCacheMiss
			}
		default:
			keyErr = cmd.Err()
		}
		if keyErr != nil {
			keysToErrs[k] = keyErr
		} else {
			loadedElementsCount++
		}
	}
	var byKeysErr error
	if len(keysToErrs) > 0 {
		byKeysErr = &KeyErr{keysToErrs: keysToErrs}
	}
	return b, loadedElementsCount, byKeysErr
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

type KeyErr struct {
	keysToErrs map[string]error
}

func (k *KeyErr) Error() string {
	return fmt.Sprintf("Load keys err: %+v", k.keysToErrs)
}

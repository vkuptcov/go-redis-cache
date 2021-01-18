package cache

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"

	"github.com/vkuptcov/go-redis-cache/v8/internal"
)

func (cd *Cache) setKV(ctx context.Context, keyValPairs ...interface{}) (err error) {
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

// Set sets multiple elements
func (cd *Cache) setMulti(ctx context.Context, items ...*Item) (err error) {
	r := cd.opt.Redis
	var pipeliner redis.Pipeliner
	if len(items) > 1 && r != nil {
		pipeliner = cd.opt.Redis.Pipeline()
		r = pipeliner
	}
	for _, item := range items {
		err = cd.setOne(ctx, r, item)
		if err != nil {
			return err
		}
	}
	if pipeliner != nil {
		_, err = pipeliner.Exec(ctx)
	}
	return err
}

func (cd *Cache) setOne(ctx context.Context, redis internal.Rediser, item *Item) error {
	value, loadValErr := item.value()
	if loadValErr != nil {
		return loadValErr
	}

	b, marshalErr := cd.Marshal(value)
	if marshalErr != nil {
		return marshalErr
	}

	ttl := cd.opt.redisTTL(item)

	if item.IfExists {
		return redis.SetXX(ctx, item.Key, b, ttl).Err()
	}

	if item.IfNotExists {
		return redis.SetNX(ctx, item.Key, b, ttl).Err()
	}

	return redis.Set(ctx, item.Key, b, ttl).Err()
}

package internal

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

func SetKV(ctx context.Context, opts *Options, keyValPairs ...interface{}) (err error) {
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
			TTL:   opts.DefaultTTL,
		}
	}
	return SetMulti(ctx, opts, items...)
}

func SetMulti(ctx context.Context, opts *Options, items ...*Item) (err error) {
	r := opts.Redis
	var pipeliner redis.Pipeliner
	if len(items) > 1 && r != nil {
		pipeliner = opts.Redis.Pipeline()
		r = pipeliner
	}
	for _, item := range items {
		err = setOne(ctx, opts, r, item)
		if err != nil {
			return err
		}
	}
	if pipeliner != nil {
		_, err = pipeliner.Exec(ctx)
	}
	return err
}

func setOne(ctx context.Context, opts *Options, redis Rediser, item *Item) error {
	b, marshalErr := opts.Marshaller.Marshal(item.Value)
	if marshalErr != nil {
		return marshalErr
	}

	ttl := opts.redisTTL(item)

	if item.IfExists {
		return redis.SetXX(ctx, item.Key, b, ttl).Err()
	}

	if item.IfNotExists {
		return redis.SetNX(ctx, item.Key, b, ttl).Err()
	}

	return redis.Set(ctx, item.Key, b, ttl).Err()
}
package internal

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

func SetKV(ctx context.Context, opts Options, keyValPairs ...interface{}) error {
	if len(keyValPairs)%2 != 0 {
		return errors.Wrapf(ErrKeyPairs, "even keyValPairs number expected, %d given", len(keyValPairs))
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

func SetMulti(ctx context.Context, opts Options, items ...*Item) (err error) {
	if len(items) == 0 {
		return nil
	}
	r := opts.Redis
	var pipeliner redis.Pipeliner
	if len(items) > 1 || items[0].Field != "" {
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

func HSetKV(ctx context.Context, opts Options, key string, fieldValPairs ...interface{}) error {
	if len(fieldValPairs)%2 != 0 {
		return ErrKeyPairs
	}
	fieldMarshalledValsPairs := make([]interface{}, len(fieldValPairs))
	for idx := 0; idx < len(fieldValPairs); idx += 2 {
		field, ok := fieldValPairs[idx].(string)
		if !ok {
			return errors.Wrapf(ErrNonStringKey, "string field expected for position %d, `%#+v` of type %T given", idx, fieldValPairs[idx], fieldValPairs[idx])
		}
		marshalledBytes, marshalErr := opts.Marshaller.Marshal(fieldValPairs[idx+1])
		if marshalErr != nil {
			return marshalErr
		}
		fieldMarshalledValsPairs[idx] = field
		fieldMarshalledValsPairs[idx+1] = string(marshalledBytes)
	}
	pipeline := opts.Redis.Pipeline()
	pipeline.HSet(ctx, key, fieldMarshalledValsPairs...)
	pipeline.Expire(ctx, key, opts.DefaultTTL)
	_, pipelineErr := pipeline.Exec(ctx)
	return pipelineErr
}

func setOne(ctx context.Context, opts Options, rediser Rediser, item *Item) error {
	b, marshalErr := opts.Marshaller.Marshal(item.Value)
	if marshalErr != nil {
		return marshalErr
	}

	ttl := opts.redisTTL(item.TTL)

	if item.Field == "" {

		if item.IfExists {
			return rediser.SetXX(ctx, item.Key, b, ttl).Err()
		}

		if item.IfNotExists {
			return rediser.SetNX(ctx, item.Key, b, ttl).Err()
		}

		return rediser.Set(ctx, item.Key, b, ttl).Err()
	} else {
		if item.IfNotExists {
			rediser.HSetNX(ctx, item.Key, item.Field, string(b))
		} else {
			rediser.HSet(ctx, item.Key, item.Field, string(b))
		}
		rediser.Expire(ctx, item.Key, ttl)
	}
	return nil
}

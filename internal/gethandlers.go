package internal

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"

	"github.com/vkuptcov/go-redis-cache/v8/internal/containers"
)

func Get(ctx context.Context, opts Options, dst interface{}, keys []string) error {
	return getInternal(ctx, opts, dst, func(pipeliner redis.Pipeliner) {
		for _, k := range keys {
			_ = pipeliner.Get(ctx, k)
		}
	})
}

func HGetAll(ctx context.Context, opts Options, dst interface{}, keys []string) error {
	return getInternal(ctx, opts, dst, func(pipeliner redis.Pipeliner) {
		for _, k := range keys {
			pipeliner.HGetAll(ctx, k)
		}
	})
}

func HGetFields(ctx context.Context, opts Options, dst interface{}, keysToFields map[string][]string) error {
	return getInternal(ctx, opts, dst, func(pipeliner redis.Pipeliner) {
		for key, fields := range keysToFields {
			pipeliner.HMGet(ctx, key, fields...)
		}
	})
}

func getInternal(ctx context.Context, opts Options, dst interface{}, pipelinerFiller func(pipeliner redis.Pipeliner)) error {
	loadErr := execAndAddIntoContainer(ctx, opts, dst, pipelinerFiller)
	if loadErr != nil && opts.AbsentKeysLoader != nil {
		var byKeyLoadErr *KeyErr
		if errors.As(loadErr, &byKeyLoadErr) && !byKeyLoadErr.HasNonCacheMissErrs() {
			absentKeys := make([]string, 0, byKeyLoadErr.CacheMissErrsCount)
			for k := range byKeyLoadErr.KeysToErrs {
				absentKeys = append(absentKeys, k)
			}
			additionallyLoadedData, additionalErr := opts.AbsentKeysLoader(absentKeys...)
			if additionalErr != nil {
				return additionalErr
			}
			return addAbsentKeys(ctx, opts, additionallyLoadedData, dst)
		}
	}
	return loadErr
}

func addAbsentKeys(ctx context.Context, opts Options, data, dst interface{}) error {
	dt := newDataTransformer(data, opts.ItemToCacheKey)
	items, transformErr := dt.getItems()
	if transformErr != nil {
		return transformErr
	}
	container, containerInitErr := containers.NewContainer(dst)
	if containerInitErr != nil {
		return containerInitErr
	}
	for _, it := range items {
		addElementToContainer(opts, container, it.Key, it.Field, it.Value)
	}
	return SetMulti(ctx, opts, items...)
}

func decodeAndAddElementToContainer(opts Options, container containers.Container, key, subkey, marshalledVal string) error {
	dstEl := container.DstEl()
	unmarshalErr := opts.Marshaller.Unmarshal([]byte(marshalledVal), dstEl)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	addElementToContainer(opts, container, key, subkey, dstEl)
	return nil
}

func addElementToContainer(opts Options, container containers.Container, key, subkey string, val interface{}) {
	if opts.CacheKeyToMapKey != nil {
		key = opts.CacheKeyToMapKey(key)
	}
	container.AddElementWithSubkey(key, subkey, val)
}

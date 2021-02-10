package internal

import (
	"context"

	"github.com/pkg/errors"

	"github.com/vkuptcov/go-redis-cache/v8/internal/containers"
)

func Get(ctx context.Context, opts Options, dst interface{}, keys []string) error {
	if opts.AbsentKeysLoader != nil {
		opts.AddCacheMissErrors = true
	}
	loadErr := getFromCache(ctx, opts, dst, keys)
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
			return addAbsentKeys(ctx, opts, additionallyLoadedData, dst, opts.ItemToCacheKey)
		}
	}
	return loadErr
}

func getFromCache(ctx context.Context, opts Options, dst interface{}, keys []string) error {
	pipeliner := opts.Redis.Pipeline()
	for _, k := range keys {
		_ = pipeliner.Get(ctx, k)
	}
	return execAndAddIntoContainer(ctx, opts, dst, pipeliner)
}

func addAbsentKeys(ctx context.Context, opts Options, data interface{}, dst interface{}, itemToKeyFn func(it interface{}) string) error {
	dt := newDataTransformer(data, itemToKeyFn)
	items, transformErr := dt.getItems()
	if transformErr != nil {
		return transformErr
	}
	container, containerInitErr := containers.NewContainer(dst)
	if containerInitErr != nil {
		return containerInitErr
	}
	for _, it := range items {
		addElementToContainer(opts, container, it.Key, it.Value)
	}
	return SetMulti(ctx, opts, items...)
}

func decodeAndAddElementToContainer(opts Options, container containers.Container, key, marshalledVal string) error {
	dstEl := container.DstEl()
	unmarshalErr := opts.Marshaller.Unmarshal([]byte(marshalledVal), dstEl)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	addElementToContainer(opts, container, key, dstEl)
	return nil
}

func addElementToContainer(opts Options, container containers.Container, key string, val interface{}) {
	if opts.CacheKeyToMapKey != nil {
		key = opts.CacheKeyToMapKey(key)
	}
	container.AddElement(key, val)
}

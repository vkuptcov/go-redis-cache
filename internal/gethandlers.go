package internal

import (
	"context"

	"github.com/go-redis/redis/v8"
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
	loadedBytes, loadedElementsCount, loadErr := getBytes(ctx, opts, keys)
	if loadErr != nil {
		return loadErr
	}

	container, containerErr := containers.NewContainer(dst)

	if errors.Is(containerErr, containers.ErrNonContainerType) {
		return opts.Marshaller.Unmarshal(loadedBytes[0], dst)
	} else if containerErr != nil {
		return containerErr
	}

	container.InitWithSize(loadedElementsCount)
	for idx, b := range loadedBytes {
		decodeErr := decodeAndAddElementToContainer(opts, container, keys[idx], string(b))
		if decodeErr != nil {
			// @todo init and add KeyErr
			return decodeErr
		}
	}
	return nil
}

// @todo optimize it for single key
func getBytes(ctx context.Context, opts Options, keys []string) (b [][]byte, loadedElementsCount int, err error) {
	pipeliner := opts.Redis.Pipeline()
	for _, k := range keys {
		_ = pipeliner.Get(ctx, k)
	}

	// errors are handled by keys
	cmds, _ := pipeliner.Exec(ctx)

	b = make([][]byte, len(keys))

	keysToErrs := map[string]error{}
	var cacheMissErrsCount int

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
			if opts.AddCacheMissErrors {
				keyErr = ErrCacheMiss
				cacheMissErrsCount++
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
		byKeysErr = &KeyErr{
			KeysToErrs:         keysToErrs,
			CacheMissErrsCount: cacheMissErrsCount,
		}
	}
	return b, loadedElementsCount, byKeysErr
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

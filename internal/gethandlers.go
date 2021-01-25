package internal

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"

	"github.com/vkuptcov/go-redis-cache/v8/internal/containers"
)

type GetLoadArgs struct {
	Dst         interface{}
	KeysToGet   []string
	LoadFn      func(absentKeys ...string) (interface{}, error)
	ItemToGetFn func(it interface{}) string
}

func Get(ctx context.Context, opts *Options, dst interface{}, keys []string) error {
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
		dstEl := container.DstEl()
		unmarshalErr := opts.Marshaller.Unmarshal(b, dstEl)
		if unmarshalErr != nil {
			// @todo init and add KeyErr
			return unmarshalErr
		}
		container.AddElement(keys[idx], dstEl)
	}
	return nil
}

func GetOrLoad(ctx context.Context, opts *Options, dst interface{}, loadFn func(absentKeys ...string) (interface{}, error), keys []string) error {
	ctx = WithCacheMissErrorsContext(ctx)
	loadErr := Get(ctx, opts, dst, keys)
	if loadErr != nil {
		var byKeyLoadErr *KeyErr
		if errors.As(loadErr, &byKeyLoadErr) && !byKeyLoadErr.HasNonCacheMissErrs() {
			absentKeys := make([]string, 0, byKeyLoadErr.CacheMissErrsCount)
			for k := range byKeyLoadErr.KeysToErrs {
				absentKeys = append(absentKeys, k)
			}
			return addAbsentKeys(ctx, opts, dst, loadFn, absentKeys)
		}
	}
	return loadErr
}

// @todo optimize it for single key
func getBytes(ctx context.Context, opts *Options, keys []string) (b [][]byte, loadedElementsCount int, err error) {
	includeCacheMissErrors, _ := ctx.Value(IncludeCacheMissErrsKey).(bool)
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
			if includeCacheMissErrors {
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

func addAbsentKeys(ctx context.Context, opts *Options, dst interface{}, loadFn func(absentKeys ...string) (interface{}, error), absentKeys []string) error {
	data, loadErr := loadFn(absentKeys...)
	if loadErr != nil {
		return loadErr
	}
	dt := newDataTransformer(data)
	items, transformErr := dt.getItems()
	if transformErr != nil {
		return transformErr
	}
	container, containerInitErr := containers.NewContainer(dst)
	if containerInitErr != nil {
		return containerInitErr
	}
	for _, it := range items {
		container.AddElement(it.Key, it.Value)
	}
	return SetMulti(ctx, opts, items...)
}

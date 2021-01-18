package cache

import (
	"context"
	"reflect"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"

	"github.com/vkuptcov/go-redis-cache/v8/internal/containers"
)

func (cd *Cache) get(ctx context.Context, dst interface{}, keys []string) error {
	loadedBytes, loadedElementsCount, loadErr := cd.getBytes(ctx, keys)

	if len(keys) == 1 {
		if loadErr != nil {
			return loadErr
		}
		return cd.Unmarshal(loadedBytes[0], dst)
	}
	container, containerErr := containers.NewContainer(dst)
	if containerErr != nil {
		return containerErr
	}
	container.InitWithSize(loadedElementsCount)
	for idx, b := range loadedBytes {
		dstEl := container.DstEl()
		unmarshalErr := cd.Unmarshal(b, dstEl)
		if unmarshalErr != nil {
			// @todo init and add KeyErr
			return unmarshalErr
		}
		container.AddElement(keys[idx], dstEl)
	}
	return nil
}

func (cd *Cache) getOrLoad(ctx context.Context, dst interface{}, loadFn func(absentKeys ...string) (interface{}, error), keys ...string) error {
	ctx = WithCacheMissErrorsContext(ctx)
	loadErr := cd.Get(ctx, dst, keys...)
	if loadErr != nil {
		var byKeyLoadErr *KeyErr
		if errors.As(loadErr, &byKeyLoadErr) && !byKeyLoadErr.HasNonCacheMissErrs() {
			absentKeys := make([]string, 0, byKeyLoadErr.cacheMissErrsCount)
			for k := range byKeyLoadErr.keysToErrs {
				absentKeys = append(absentKeys, k)
			}
			return cd.addAbsentKeys(ctx, dst, loadFn, absentKeys)
		}
	}
	return loadErr
}

// @todo optimize it for single key
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
			keysToErrs:         keysToErrs,
			cacheMissErrsCount: cacheMissErrsCount,
		}
	}
	return b, loadedElementsCount, byKeysErr
}

func (cd *Cache) addAbsentKeys(ctx context.Context, dst interface{}, loadFn func(absentKeys ...string) (interface{}, error), absentKeys []string) error {
	data, loadErr := loadFn(absentKeys...)
	if loadErr != nil {
		return loadErr
	}
	v := reflect.ValueOf(data)
	switch kind := v.Kind(); kind {
	case reflect.Map:
		if v.Len() == 0 {
			return nil
		}
		mapType := v.Type()
		keyType := mapType.Key()
		if keyType.Kind() != reflect.String {
			return errors.Errorf("dst key type must be a string, %v given", keyType.Kind())
		}
		container, containerInitErr := containers.NewContainer(dst)
		if containerInitErr != nil {
			return containerInitErr
		}
		iter := v.MapRange()
		items := make([]*Item, 0, v.Len())
		for iter.Next() {
			val := iter.Value().Interface()
			var key string
			if item, ok := val.(*Item); ok {
				items = append(items, item)
				// @todo add possibility to use the key from the map
				key = item.Key
			} else {
				key = iter.Key().String()
				items = append(items, &Item{
					Key:   key,
					Value: val,
				})
			}
			container.AddElement(key, val)
		}
		return cd.Set(ctx, items...)
	default:
		return errors.Errorf("Unsupported kind %q", kind)
	}
}

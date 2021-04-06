package internal

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"

	"github.com/vkuptcov/go-redis-cache/v8/internal/containers"
)

func execAndAddIntoContainer(ctx context.Context, opts Options, dst interface{}, pipelinerFiller func(pipeliner redis.Pipeliner)) error {
	if opts.AbsentKeysLoader != nil {
		opts.AddCacheMissErrors = true
	}
	pipeliner := opts.Redis.Pipeline()
	pipelinerFiller(pipeliner)

	// pipeliner errs will be checked for all the keys
	cmds, _ := pipeliner.Exec(ctx)

	container, containerInitErr := containers.NewContainer(dst)
	if containerInitErr != nil {
		return containerInitErr
	}
	container.InitWithSize(len(cmds))

	byKeysErr := &KeyErr{
		KeysToErrs:         map[string]error{},
		CacheMissErrsCount: 0,
	}

	for _, cmderr := range cmds {
		key := cmderr.Args()[1].(string)
		if cmderr.Err() != nil {
			if errors.Is(cmderr.Err(), redis.Nil) {
				if opts.AddCacheMissErrors {
					byKeysErr.AddErrorForKey(key, errors.Wrapf(ErrCacheMiss, "key %q not found", key))
				}
			} else {
				byKeysErr.AddErrorForKey(key, cmderr.Err())
			}
			continue
		}

		switch typedCmd := cmderr.(type) {
		// returned for HMGET
		case *redis.SliceCmd:
			handleSliceCmd(opts, typedCmd, container, key, byKeysErr)
		// returned for HGETALL
		case *redis.StringStringMapCmd:
			handleStringStringMapCmd(opts, typedCmd, container, key, byKeysErr)
		case *redis.StringCmd:
			handleStringCmd(opts, typedCmd, container, key, byKeysErr)
		}
	}
	if len(byKeysErr.KeysToErrs) > 0 {
		return byKeysErr
	}
	return nil
}

func handleSliceCmd(opts Options, typedCmd *redis.SliceCmd, container containers.Container, key string, byKeysErr *KeyErr) {
	fields := typedCmd.Args()[2:]
	for fieldIdx, val := range typedCmd.Val() {
		field := fields[fieldIdx].(string)
		switch t := val.(type) {
		case error:
			if errors.Is(t, redis.Nil) {
				if opts.AddCacheMissErrors {
					byKeysErr.AddErrorForKeyAndField(key, field, errors.Wrapf(ErrCacheMiss, "key %q and field %q not found", key, field))
				}
			} else {
				byKeysErr.AddErrorForKeyAndField(key, field, t)
			}
		case string:
			if decodeErr := decodeAndAddElementToContainer(opts, container, key, field, t); decodeErr != nil {
				byKeysErr.AddErrorForKeyAndField(key, field, decodeErr)
			}
		default:
			if t == nil && opts.AddCacheMissErrors {
				byKeysErr.AddErrorForKeyAndField(key, field, errors.Wrapf(ErrCacheMiss, "key %q and field %q not found", key, field))
			} else {
				byKeysErr.AddErrorForKeyAndField(key, field, errors.Errorf("Non-handled type returned: %T", t))
			}
		}
	}
}

func handleStringStringMapCmd(opts Options, typedCmd *redis.StringStringMapCmd, container containers.Container, key string, byKeysErr *KeyErr) {
	for field, val := range typedCmd.Val() {
		decodeErr := decodeAndAddElementToContainer(opts, container, key, field, val)
		if decodeErr != nil {
			byKeysErr.AddErrorForKeyAndField(key, field, decodeErr)
		}
	}
	// HGETALL doesn't return redis.Nil error for absent keys and returns just an empty list
	if len(typedCmd.Val()) == 0 {
		byKeysErr.AddErrorForKey(key, errors.Wrapf(ErrCacheMiss, "key %q not found", key))
	}
}

func handleStringCmd(opts Options, typedCmd *redis.StringCmd, container containers.Container, key string, byKeysErr *KeyErr) {
	decodeErr := decodeAndAddElementToContainer(opts, container, key, "", typedCmd.Val())
	if decodeErr != nil {
		byKeysErr.AddErrorForKey(key, decodeErr)
	}
}

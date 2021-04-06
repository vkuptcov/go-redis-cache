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
			if handleErr := handleSliceCmd(typedCmd, opts, container, key, byKeysErr); handleErr != nil {
				return handleErr
			}
		// returned for HGETALL
		case *redis.StringStringMapCmd:
			handleStringStringMapCmd(typedCmd, opts, container, key, byKeysErr)
		case *redis.StringCmd:
			decodeErr := decodeAndAddElementToContainer(opts, container, key, "", typedCmd.Val())
			if decodeErr != nil {
				byKeysErr.AddErrorForKey(key, decodeErr)
			}
		}
	}
	if len(byKeysErr.KeysToErrs) > 0 {
		return byKeysErr
	}
	return nil
}

func handleStringStringMapCmd(typedCmd *redis.StringStringMapCmd, opts Options, container containers.Container, key string, byKeysErr *KeyErr) {
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

func handleSliceCmd(typedCmd *redis.SliceCmd, opts Options, container containers.Container, key string, byKeysErr *KeyErr) error {
	fields := typedCmd.Args()[2:]
	for fieldIdx, val := range typedCmd.Val() {
		field := fields[fieldIdx].(string)
		switch t := val.(type) {
		case error:
			if !errors.Is(t, redis.Nil) {
				return t
			}
		case string:
			decodeErr := decodeAndAddElementToContainer(opts, container, key, field, t)
			if decodeErr != nil {
				byKeysErr.AddErrorForKeyAndField(key, field, decodeErr)
			}
		default:
			if t == nil {
				if opts.AddCacheMissErrors {
					byKeysErr.AddErrorForKeyAndField(key, field, errors.Wrapf(ErrCacheMiss, "key %q and field %q not found", key, field))
				}
				continue
			}
			return errors.Errorf("Non-handled type returned: %T", t)
		}
	}
	return nil
}

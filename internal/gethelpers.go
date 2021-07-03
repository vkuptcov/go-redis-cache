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
	isSingleElementContainer := !container.IsMultiElementContainer()
	// in case we don't have the desired element in cache and we want just to load a single one,
	// it's convenient to return a single cache miss error instead of KeyErr because the key is already known
	returnErrCacheMiss := isSingleElementContainer &&
		!opts.DisableCacheMissErrorsForSingleElementDst &&
		!opts.AddCacheMissErrors
	if returnErrCacheMiss || opts.AbsentKeysLoader != nil {
		opts.AddCacheMissErrors = true
	}

	byKeysErr := handleCmds(opts, cmds, container)

	if len(byKeysErr.KeysToErrs) > 0 {
		if returnErrCacheMiss && len(byKeysErr.KeysToErrs) == 1 && byKeysErr.CacheMissErrsCount == 1 {
			return ErrCacheMiss
		}
		return byKeysErr
	}
	return nil
}

func handleCmds(opts Options, cmds []redis.Cmder, container containers.Container) (byKeysErr *KeyErr) {
	byKeysErr = &KeyErr{
		KeysToErrs:         map[string]error{},
		CacheMissErrsCount: 0,
	}
	for _, cmderr := range cmds {
		key := cmderr.Args()[1].(string)
		if cmderr.Err() != nil {
			if errors.Is(cmderr.Err(), redis.Nil) {
				if opts.AddCacheMissErrors {
					byKeysErr.AddErrorForKey(key, ErrCacheMiss)
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
	return byKeysErr
}

func handleSliceCmd(opts Options, typedCmd *redis.SliceCmd, container containers.Container, key string, byKeysErr *KeyErr) {
	fields := typedCmd.Args()[2:]
	for fieldIdx, val := range typedCmd.Val() {
		field := fields[fieldIdx].(string)
		switch t := val.(type) {
		case error:
			if errors.Is(t, redis.Nil) {
				if opts.AddCacheMissErrors {
					byKeysErr.AddErrorForKeyAndField(key, field, ErrCacheMiss)
				}
			} else {
				byKeysErr.AddErrorForKeyAndField(key, field, t)
			}
		case string:
			if decodeErr := decodeAndAddElementToContainer(opts, container, key, field, t); decodeErr != nil {
				byKeysErr.AddErrorForKeyAndField(key, field, decodeErr)
			}
		default:
			if t == nil {
				if opts.AddCacheMissErrors {
					byKeysErr.AddErrorForKeyAndField(key, field, ErrCacheMiss)
				}
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
	if len(typedCmd.Val()) == 0 && opts.AddCacheMissErrors {
		byKeysErr.AddErrorForKey(key, ErrCacheMiss)
	}
}

func handleStringCmd(opts Options, typedCmd *redis.StringCmd, container containers.Container, key string, byKeysErr *KeyErr) {
	decodeErr := decodeAndAddElementToContainer(opts, container, key, "", typedCmd.Val())
	if decodeErr != nil {
		byKeysErr.AddErrorForKey(key, decodeErr)
	}
}

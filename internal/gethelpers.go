package internal

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"

	"github.com/vkuptcov/go-redis-cache/v8/internal/containers"
)

//nolint:gocognit,gocyclo // @todo move switch cases in different functions
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

	keysToErrs := map[string]error{}
	var cacheMissErrsCount int

	for _, cmderr := range cmds {
		key := cmderr.Args()[1].(string)
		if cmderr.Err() != nil {
			if errors.Is(cmderr.Err(), redis.Nil) {
				if opts.AddCacheMissErrors {
					keysToErrs[key] = ErrCacheMiss
					cacheMissErrsCount++
				}
			} else {
				keysToErrs[key] = cmderr.Err()
			}
			continue
		}

		switch typedCmd := cmderr.(type) {
		case *redis.SliceCmd:
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
						// @todo init and add KeyErr
						// @todo unify with getFromCache from gethandlers
						return decodeErr
					}
				}
			}
		case *redis.StringStringMapCmd:
			for field, val := range typedCmd.Val() {
				decodeErr := decodeAndAddElementToContainer(opts, container, key, field, val)
				if decodeErr != nil {
					// @todo init and add KeyErr
					// @todo unify with getFromCache from gethandlers
					return decodeErr
				}
			}
		case *redis.StringCmd:
			decodeErr := decodeAndAddElementToContainer(opts, container, key, "", typedCmd.Val())
			if decodeErr != nil {
				// @todo init and add KeyErr
				// @todo unify with getFromCache from gethandlers
				return decodeErr
			}
		}
	}
	var byKeysErr error
	if len(keysToErrs) > 0 {
		byKeysErr = &KeyErr{
			KeysToErrs:         keysToErrs,
			CacheMissErrsCount: cacheMissErrsCount,
		}
	}
	return byKeysErr
}

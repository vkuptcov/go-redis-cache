package internal

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"

	"github.com/vkuptcov/go-redis-cache/v8/internal/containers"
)

func HGetAll(ctx context.Context, opts Options, dst interface{}, keys []string) error {
	pipeline := opts.Redis.Pipeline()
	for _, k := range keys {
		pipeline.HGetAll(ctx, k)
	}
	return execAndAddIntoContainer(ctx, opts, dst, pipeline)
}

func HGetFields(ctx context.Context, opts Options, dst interface{}, keysToFields map[string][]string) error {
	pipeline := opts.Redis.Pipeline()
	for key, fields := range keysToFields {
		pipeline.HMGet(ctx, key, fields...)
	}
	return execAndAddIntoContainer(ctx, opts, dst, pipeline)
}

func execAndAddIntoContainer(ctx context.Context, opts Options, dst interface{}, pipeliner redis.Pipeliner) error {
	cmds, pipelineErr := pipeliner.Exec(ctx)
	if pipelineErr != nil {
		return pipelineErr
	}
	container, containerInitErr := containers.NewContainer(dst)
	if containerInitErr != nil {
		return containerInitErr
	}
	container.InitWithSize(len(cmds))
	for _, cmderr := range cmds {
		if cmderr.Err() != nil {
			return cmderr.Err()
		}

		switch typedCmd := cmderr.(type) {
		case *redis.SliceCmd:
			key := cmderr.Args()[1].(string)
			fields := typedCmd.Args()[2:]
			for fieldIdx, val := range typedCmd.Val() {
				field := fields[fieldIdx].(string)
				switch t := val.(type) {
				case error:
					if !errors.Is(t, redis.Nil) {
						return t
					}
				case string:
					dstEl := container.DstEl()
					unmarshalErr := opts.Marshaller.Unmarshal([]byte(t), dstEl)
					if unmarshalErr != nil {
						// @todo init and add KeyErr
						// @todo unify with getFromCache from gethandlers
						return unmarshalErr
					}
					addElementToContainer(opts, container, key+"-"+field, dstEl)
				}
			}
		case *redis.StringStringMapCmd:
			key := cmderr.Args()[1].(string)
			for field, val := range typedCmd.Val() {
				dstEl := container.DstEl()
				unmarshalErr := opts.Marshaller.Unmarshal([]byte(val), dstEl)
				if unmarshalErr != nil {
					// @todo init and add KeyErr
					// @todo unify with getFromCache from gethandlers
					return unmarshalErr
				}
				addElementToContainer(opts, container, key+"-"+field, dstEl)
			}
		}
	}
	return nil
}

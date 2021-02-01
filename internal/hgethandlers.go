package internal

import (
	"context"

	"github.com/go-redis/redis/v8"

	"github.com/vkuptcov/go-redis-cache/v8/internal/containers"
)

func HGetAll(ctx context.Context, opts Options, dst interface{}, keys []string) error {
	pipeline := opts.Redis.Pipeline()
	for _, k := range keys {
		pipeline.HGetAll(ctx, k)
	}
	cmds, pipelineErr := pipeline.Exec(ctx)
	if pipelineErr != nil {
		return pipelineErr
	}
	container, containerInitErr := containers.NewContainer(dst)
	if containerInitErr != nil {
		return containerInitErr
	}
	for idx, cmderr := range cmds {
		if cmderr.Err() != nil {
			return cmderr.Err()
		}
		key := keys[idx]
		if stringStringCmd, ok := cmderr.(*redis.StringStringMapCmd); ok {
			for field, val := range stringStringCmd.Val() {
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

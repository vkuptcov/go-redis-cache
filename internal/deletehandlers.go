package internal

import (
	"context"
)

func Delete(ctx context.Context, opts Options, keys []string) error {
	switch len(keys) {
	case 0:
		return nil
	case 1:
		cmd := opts.Redis.Del(keys[0])
		return cmd.Err()
	default:
		pipeliner := opts.Redis.Pipeline()
		for _, k := range keys {
			pipeliner.Del(k)
		}
		_, err := pipeliner.ExecContext(ctx)
		return err
	}
}

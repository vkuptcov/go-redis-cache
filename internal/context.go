package internal

import (
	"context"
)

type cacheContextKey int

const IncludeCacheMissErrsKey cacheContextKey = iota

func WithCacheMissErrorsContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, IncludeCacheMissErrsKey, true)
}

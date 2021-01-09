package cache

import "context"

type cacheContextKey int

const includeCacheMissErrsKey cacheContextKey = iota

func WithCacheMissErrorsContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, includeCacheMissErrsKey, true)
}

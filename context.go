package cache

import (
	"context"

	"github.com/vkuptcov/go-redis-cache/v8/internal"
)

func WithCacheMissErrorsContext(ctx context.Context) context.Context {
	return internal.WithCacheMissErrorsContext(ctx)
}

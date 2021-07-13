package cache

import (
	"github.com/vkuptcov/go-redis-cache/v7/internal"
)

type Item = internal.Item

type Options = internal.Options

type KeyErr = internal.KeyErr

var ErrItemToCacheKeyFnRequired = internal.ErrItemToCacheKeyFnRequired
var ErrCacheMiss = internal.ErrCacheMiss

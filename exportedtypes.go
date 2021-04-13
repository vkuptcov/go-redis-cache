package cache

import (
	"github.com/vkuptcov/go-redis-cache/v8/internal"
)

type Marshaller internal.Marshaller

type Item = internal.Item

type Options = internal.Options

type KeyErr = internal.KeyErr

var ErrItemToCacheKeyFnRequired = internal.ErrItemToCacheKeyFnRequired
var ErrCacheMiss = internal.ErrCacheMiss

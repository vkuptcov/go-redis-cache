package cache

import (
	"github.com/vkuptcov/go-redis-cache/v8/internal"
	"github.com/vkuptcov/go-redis-cache/v8/marshallers"
)

type Marshaller marshallers.Marshaller

type Item = internal.Item

type Options = internal.Options

type KeyErr = internal.KeyErr

var ErrItemToCacheKeyFnRequired = internal.ErrItemToCacheKeyFnRequired
var ErrCacheMiss = internal.ErrCacheMiss

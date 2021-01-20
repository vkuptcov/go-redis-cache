package cache

import (
	"time"

	"github.com/vkuptcov/go-redis-cache/v8/internal"
)

type Options internal.Options

func (opt *Options) init() {
	cacheDuration := time.Hour
	if opt.DefaultTTL >= 1*time.Second {
		cacheDuration = opt.DefaultTTL
	}
	opt.DefaultTTL = cacheDuration
}

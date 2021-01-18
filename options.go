package cache

import (
	"time"
)

type Options struct {
	Redis rediser

	// DefaultTTL is the cache expiration time.
	// 1 hour by default
	DefaultTTL time.Duration

	Marshaller Marshaller
}

func (opt *Options) init() {
	cacheDuration := time.Hour
	if opt.DefaultTTL >= 1*time.Second {
		cacheDuration = opt.DefaultTTL
	}
	opt.DefaultTTL = cacheDuration
}

package internal

import "time"

type Options struct {
	Redis Rediser

	// DefaultTTL is the cache expiration time.
	// 1 hour by default
	DefaultTTL time.Duration

	Marshaller Marshaller
}

func (opt Options) redisTTL(item *Item) time.Duration {
	if item.TTL < 0 {
		return 0
	}
	if item.TTL < time.Second {
		return opt.DefaultTTL
	}
	return item.TTL
}

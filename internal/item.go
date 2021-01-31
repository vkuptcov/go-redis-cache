package internal

import (
	"time"
)

type Item struct {
	Key string

	// Value to be cached
	Value interface{}

	// TTL is the cache expiration time.
	// Default TTL is taken from Options
	TTL time.Duration

	// IfExists only sets the key if it already exist.
	IfExists bool

	// IfNotExists only sets the key if it does not already exist.
	// Only one of IfExists/IfNotExists can be setOne
	IfNotExists bool
}

type HItem struct {
	Key string

	Field string

	// Value to be cached
	Value interface{}

	// TTL is the cache expiration time.
	// Default TTL is taken from Options
	TTL time.Duration

	// IfNotExists only sets the key if it does not already exist.
	// Only one of IfExists/IfNotExists can be setOne
	IfNotExists bool
}

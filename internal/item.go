package internal

import (
	"time"
)

// Item represents all the data needed to cache an element
type Item struct {
	// Key represents a Redis key which will be used to cache value
	Key string

	// Field is an optional parameter. If it's set than Redis hash maps will be used
	// it make the API a bit less cohesive, but actually simplifies the implementation and usage,
	// and allows to mix different items together if needed
	Field string

	// Value to be cached
	Value interface{}

	// TTL is the cache expiration time.
	// Default TTL is taken from Options
	TTL time.Duration

	// IfExists only sets the key if it already exist.
	// Doesn't work if Field is set: Redis hash maps doesn't support it
	IfExists bool

	// IfNotExists only sets the key if it does not already exist.
	// Only one of IfExists/IfNotExists can be setOne
	IfNotExists bool
}

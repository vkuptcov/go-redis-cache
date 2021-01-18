package cache

import "time"

type Item struct {
	Key string

	// Value to be cached
	Value interface{}

	// Load returns value to be cached.
	Load func(*Item) (interface{}, error)

	// TTL is the cache expiration time.
	// Default TTL is taken from Options
	TTL time.Duration

	// IfExists only sets the key if it already exist.
	IfExists bool

	// IfNotExists only sets the key if it does not already exist.
	// Only one of IfExists/IfNotExists can be setOne
	IfNotExists bool
}

func (item *Item) value() (interface{}, error) {
	if item.Value != nil {
		return item.Value, nil
	}
	if item.Load != nil {
		val, err := item.Load(item)
		item.Value = val
		return val, err
	}
	return nil, nil
}

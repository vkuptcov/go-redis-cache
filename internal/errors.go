package internal

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/vkuptcov/go-redis-cache/v8/cachekeys"
)

var ErrCacheMiss = errors.New("cache: key is missing")
var ErrWrongLoadFnType = errors.New("load function must return slice or key-value map")
var ErrKeyPairs = errors.New("key-values pairs must be provided")
var ErrNonStringKey = errors.New("string key expected")
var ErrItemToCacheKeyFnRequired = errors.New("CacheKeyExtractor transformation function must be set or only *Item's can be returned from the loader function")

type KeyErr struct {
	KeysToErrs         map[string]error
	CacheMissErrsCount int
}

func (k *KeyErr) Error() string {
	return fmt.Sprintf("Load keys err: %+v", k.KeysToErrs)
}

func (k *KeyErr) HasNonCacheMissErrs() bool {
	return len(k.KeysToErrs) > k.CacheMissErrsCount
}

func (k *KeyErr) AddErrorForKey(key string, err error) {
	prevErr := k.KeysToErrs[key]
	if prevErr == nil {
		k.KeysToErrs[key] = errors.Wrapf(err, "Key %q load failed", key)
		if errors.Is(err, ErrCacheMiss) {
			k.CacheMissErrsCount++
		}
	}
}

func (k *KeyErr) AddErrorForKeyAndField(key, field string, err error) {
	keyWithField := cachekeys.KeyWithField(key, field)
	prevErr := k.KeysToErrs[keyWithField]
	if prevErr == nil {
		k.KeysToErrs[keyWithField] = errors.Wrapf(err, "Key %q with field %q load failed", key, field)
		if errors.Is(err, ErrCacheMiss) {
			k.CacheMissErrsCount++
		}
	}
}

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
var ErrItemToCacheKeyFnRequired = errors.New("ItemToCacheKey transformation function must be set or only *Item's can be returned from the loader function")

type KeyErr struct {
	KeysToErrs         map[string]error
	CacheMissErrsCount int
}

func (k *KeyErr) Error() string {
	return fmt.Sprintf("Load keys err: %+v", k.KeysToErrs)
}

func (k *KeyErr) HasNonCacheMissErrs() bool {
	return len(k.KeysToErrs) != k.CacheMissErrsCount
}

func (k *KeyErr) AddErrorForKey(key string, err error) {
	k.KeysToErrs[key] = err
	if errors.Is(err, ErrCacheMiss) {
		k.CacheMissErrsCount++
	}
}

func (k *KeyErr) AddErrorForKeyAndField(key, field string, err error) {
	k.KeysToErrs[cachekeys.KeyWithField(key, field)] = err
	if errors.Is(err, ErrCacheMiss) {
		k.CacheMissErrsCount++
	}
}

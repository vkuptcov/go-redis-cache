package cache

import (
	"fmt"

	"github.com/pkg/errors"
)

var ErrCacheMiss = errors.New("cache: key is missing")

type KeyErr struct {
	keysToErrs         map[string]error
	cacheMissErrsCount int
}

func (k *KeyErr) HasNonCacheMissErrs() bool {
	return len(k.keysToErrs) != k.cacheMissErrsCount
}

func (k *KeyErr) Error() string {
	return fmt.Sprintf("Load keys err: %+v", k.keysToErrs)
}

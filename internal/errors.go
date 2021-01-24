package internal

import (
	"fmt"

	"github.com/pkg/errors"
)

var ErrCacheMiss = errors.New("cache: key is missing")

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
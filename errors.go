package cache

import (
	"github.com/vkuptcov/go-redis-cache/v8/internal"
)

type KeyErr internal.KeyErr

func (k *KeyErr) Error() string {
	return (*internal.KeyErr)(k).Error()
}

func (k *KeyErr) HasNonCacheMissErrs() bool {
	return (*internal.KeyErr)(k).HasNonCacheMissErrs()
}

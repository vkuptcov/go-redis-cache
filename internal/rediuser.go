package internal

import (
	"time"

	"github.com/go-redis/redis/v7"
)

type Rediser interface {
	Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	SetXX( key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	SetNX( key string, value interface{}, expiration time.Duration) *redis.BoolCmd

	Get( key string) *redis.StringCmd

	HMGet( key string, fields ...string) *redis.SliceCmd
	HGetAll( key string) *redis.StringStringMapCmd

	HSet( key string, values ...interface{}) *redis.IntCmd
	HSetNX( key, field string, value interface{}) *redis.BoolCmd

	Expire( key string, expiration time.Duration) *redis.BoolCmd

	Del( keys ...string) *redis.IntCmd

	Pipeline() redis.Pipeliner
}

var _ Rediser = redis.UniversalClient(nil)

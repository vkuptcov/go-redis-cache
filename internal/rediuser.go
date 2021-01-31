package internal

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type Rediser interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	SetXX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd

	Get(ctx context.Context, key string) *redis.StringCmd

	HMGet(ctx context.Context, key string, fields ...string) *redis.SliceCmd
	HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd

	HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	HSetNX(ctx context.Context, key, field string, value interface{}) *redis.BoolCmd

	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd

	Del(ctx context.Context, keys ...string) *redis.IntCmd

	Pipeline() redis.Pipeliner
}

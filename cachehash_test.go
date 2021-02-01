package cache_test

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/suite"
	"syreclabs.com/go/faker"

	cache "github.com/vkuptcov/go-redis-cache/v8"
	"github.com/vkuptcov/go-redis-cache/v8/internal/marshaller"
)

type CacheHashSuite struct {
	suite.Suite
	client *redis.Client
	cache  *cache.Cache
}

func (st *CacheHashSuite) SetupSuite() {
	st.client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	st.cache = cache.NewCache(cache.Options{
		Redis:      st.client,
		DefaultTTL: 0,
		Marshaller: marshaller.NewMarshaller(&marshaller.JSONMarshaller{}),
	})
}

func (st *CacheHashSuite) TestHSet() {
	ctx := context.Background()

	key := faker.RandomString(10)

	expectedMap := map[string]string{}
	var fieldValPairs []interface{}
	for i := 0; i < 3; i++ {
		field := faker.RandomString(5)
		val := faker.Lorem().Word()
		fieldValPairs = append(fieldValPairs, field, val)
		expectedMap[field] = val
	}
	hsetErr := st.cache.HSetKV(ctx, key, fieldValPairs...)
	st.Require().NoError(hsetErr, "no error expected on hset")

	stringMapCmd := st.client.HGetAll(ctx, key)
	st.Require().NoError(stringMapCmd.Err(), "No error expected on getting hash from redis directly")

	st.Require().Empty(cmp.Diff(expectedMap, stringMapCmd.Val()), "unmatched result")
}

func (st *CacheHashSuite) TestHGetAll() {
	ctx := context.Background()

	var expected []string

	keys := []string{faker.RandomString(10), faker.RandomString(10)}

	for _, k := range keys {
		var fieldValPairs []interface{}
		for i := 0; i < 3; i++ {
			field := faker.RandomString(5)
			val := faker.Lorem().Word()
			fieldValPairs = append(fieldValPairs, field, val)
			expected = append(expected, val)
		}
		hsetErr := st.cache.HSetKV(ctx, k, fieldValPairs...)
		st.Require().NoError(hsetErr, "no error expected on hset")
	}

	var dst []string
	hgetErr := st.cache.HGetAll(ctx, &dst, keys...)
	st.Require().NoError(hgetErr, "No error expected on HGetAll")

	st.Require().ElementsMatch(expected, dst, "unexpected slice")
}

func TestCacheHashSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &CacheHashSuite{})
}

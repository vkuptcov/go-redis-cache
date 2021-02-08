package cache_test

import (
	"context"
	"strings"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/suite"
	"syreclabs.com/go/faker"

	cache "github.com/vkuptcov/go-redis-cache/v8"
	"github.com/vkuptcov/go-redis-cache/v8/internal/marshaller"
)

type CacheHashSuite struct {
	suite.Suite
	client *redis.Client
	cache  *cache.Cache

	keys          []string
	keysToFields  map[string][]string
	expectedSlice []string
	expectedMap   map[string]string
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

func (st *CacheHashSuite) SetupTest() {
	st.keys = []string{faker.RandomString(10), faker.RandomString(10)}
	st.expectedMap = map[string]string{}
	st.keysToFields = map[string][]string{}

	for _, k := range st.keys {
		var fieldValPairs []interface{}
		for i := 0; i < 3; i++ {
			field := faker.RandomString(5)
			val := faker.Lorem().Word()
			fieldValPairs = append(fieldValPairs, field, val)
			st.expectedSlice = append(st.expectedSlice, val)
			st.expectedMap[k+"-"+field] = val
			st.keysToFields[k] = append(st.keysToFields[k], field)
		}
		hsetErr := st.cache.HSetKV(context.Background(), k, fieldValPairs...)
		st.Require().NoError(hsetErr, "no error expectedSlice on hset")
	}
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

	checkDst(st.T(), expectedMap, stringMapCmd.Val(), "unmatched result")
}

func (st *CacheHashSuite) TestHGetAll() {
	ctx := context.Background()

	var sliceDst []string
	hgetErr := st.cache.HGetAll(ctx, &sliceDst, st.keys...)
	st.Require().NoError(hgetErr, "No error expectedSlice on HGetAll")
	st.Require().ElementsMatch(st.expectedSlice, sliceDst, "unexpected slice")

	var mapDst map[string]string
	hgetErr = st.cache.HGetAll(ctx, &mapDst, st.keys...)
	st.Require().NoError(hgetErr, "No error expectedSlice on HGetAll")
	checkDst(st.T(), st.expectedMap, mapDst, "unmatched result")
}

func (st *CacheHashSuite) TestHGetFields() {
	for key, fields := range st.keysToFields {
		st.Run("load "+key+" into slice", func() {
			var dst []string
			hgetErr := st.cache.HGetFieldsForKey(context.Background(), &dst, key, fields...)
			st.Require().NoError(hgetErr, "no error expected on HGetFieldsForKey")
			var expectedSlice []string
			for k, val := range st.expectedMap {
				if strings.HasPrefix(k, key) {
					expectedSlice = append(expectedSlice, val)
				}
			}
			checkDst(st.T(), expectedSlice, dst, "unexpected slice")
		})
		st.Run("load "+key+" into map", func() {
			var dst map[string]string
			hgetErr := st.cache.HGetFieldsForKey(context.Background(), &dst, key, fields...)
			st.Require().NoError(hgetErr, "no error expected on HGetFieldsForKey")
			expectedMap := map[string]string{}
			for k, val := range st.expectedMap {
				if strings.HasPrefix(k, key) {
					expectedMap[k] = val
				}
			}
			checkDst(st.T(), expectedMap, dst, "unexpected map")
		})
	}

}

func TestCacheHashSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &CacheHashSuite{})
}

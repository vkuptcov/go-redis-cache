package cache_test

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v7"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"syreclabs.com/go/faker"

	cache "github.com/vkuptcov/go-redis-cache/v7"
	"github.com/vkuptcov/go-redis-cache/v7/marshallers"
)

type BaseCacheSuite struct {
	suite.Suite
	client     *redis.Client
	cache      *cache.Cache
	marshaller marshallers.Marshaller
	ctx        context.Context
}

func (st *BaseCacheSuite) SetupSuite() {
	st.client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	st.ctx = context.Background()

	st.marshaller = marshallers.NewMarshaller(&marshallers.JSONMarshaller{})

	st.cache = cache.NewCache(cache.Options{
		Redis:      st.client,
		DefaultTTL: 0,
		Marshaller: st.marshaller,
	})
}

func (st *BaseCacheSuite) keysToMap(keys ...string) map[string]string {
	m := map[string]string{}
	for _, k := range keys {
		m[k] = st.keyToElement(k)
	}
	return m
}

func (st *BaseCacheSuite) keyToElement(k string) string {
	return k + "-element"
}

func (st *BaseCacheSuite) checkElementsInCache(expected map[string]string) {
	st.T().Helper()
	var keys []string
	for k := range expected {
		keys = append(keys, k)
	}
	var dst map[string]string
	st.Require().NoError(
		st.cache.Get(context.Background(), &dst, keys...),
		"No error expected on checking keysToLoad in cache",
	)
	checkDst(st.T(), expected, dst, "difference in cache found")
}

func (st *BaseCacheSuite) generateKeyValPairs() (data commonTestData) {
	data.keyVals = map[string]string{}
	for i := 0; i < 3; i++ {
		key := faker.RandomString(5)
		val := faker.Lorem().Word()
		data.keyValPairs = append(data.keyValPairs, key, val)
		data.keys = append(data.keys, key)
		data.vals = append(data.vals, val)
		data.keyVals[key] = val
	}
	return data
}

type commonTestData struct {
	keyVals     map[string]string
	keyValPairs []interface{}
	keys        []string
	vals        []string
}

func checkDst(t *testing.T, expected, actual interface{}, msgAndArs ...interface{}) {
	t.Helper()
	diff := cmp.Diff(
		expected,
		actual,
		cmpopts.SortSlices(func(x, y string) bool {
			return x > y
		}),
		cmpopts.SortMaps(func(x, y string) bool {
			return x > y
		}),
	)
	require.New(t).Empty(diff, msgAndArs...)
}

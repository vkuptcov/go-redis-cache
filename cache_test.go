package cache_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"syreclabs.com/go/faker"

	cache "github.com/vkuptcov/go-redis-cache/v8"
	"github.com/vkuptcov/go-redis-cache/v8/internal/marshaller"
)

type BaseCacheSuite struct {
	suite.Suite
	client     *redis.Client
	cache      *cache.Cache
	marshaller cache.Marshaller
	ctx        context.Context
}

func (st *BaseCacheSuite) SetupSuite() {
	st.client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	st.ctx = context.Background()

	st.marshaller = marshaller.NewMarshaller(&marshaller.JSONMarshaller{})

	st.cache = cache.NewCache(cache.Options{
		Redis:      st.client,
		DefaultTTL: 0,
		Marshaller: st.marshaller,
	})
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

type CacheSuite struct {
	suite.Suite
	client         *redis.Client
	cache          *cache.Cache
	commonTestData commonTestData
}

type commonTestData struct {
	keyVals     map[string]string
	keyValPairs []interface{}
	keys        []string
	vals        []string
}

const nonExistKey = "non-exist-key"

func (st *CacheSuite) SetupSuite() {
	st.client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	st.cache = cache.NewCache(cache.Options{
		Redis:      st.client,
		DefaultTTL: 0,
		Marshaller: marshaller.NewMarshaller(&marshaller.JSONMarshaller{}),
	})
}

func (st *CacheSuite) SetupTest() {
	keyVals := map[string]string{}
	var keyValPairs []interface{}
	var keys []string
	var vals []string
	for i := 0; i < 5; i++ {
		k := faker.RandomString(5)
		v := faker.Lorem().Sentence(2)
		keyVals[k] = v
		keyValPairs = append(keyValPairs, k, v)
		keys = append(keys, k)
		vals = append(vals, v)
	}
	st.commonTestData = commonTestData{
		keyVals:     keyVals,
		keyValPairs: keyValPairs,
		keys:        keys,
		vals:        vals,
	}
}

func (st *CacheSuite) TestSet_DifferentItems() {
	ctx := context.Background()

	testData := []struct {
		testCase string
		items    []*cache.Item
		dst      func() interface{}
	}{
		{
			testCase: "set single item",
			items: []*cache.Item{
				{
					Key:   faker.RandomString(10),
					Value: faker.Lorem().Sentence(5),
				},
			},
			dst: func() interface{} {
				return ""
			},
		},
		{
			testCase: "set several items",
			items: []*cache.Item{
				{
					Key:   faker.RandomString(10),
					Value: faker.Lorem().Sentence(5),
				},
				{
					Key:   faker.RandomString(10),
					Value: faker.Lorem().Sentence(5),
				},
				{
					Key:   faker.RandomString(10),
					Value: faker.Lorem().Sentence(5),
				},
			},
			dst: func() interface{} {
				return ""
			},
		},
	}
	for _, td := range testData {
		st.Run(td.testCase, func() {
			setErr := st.cache.Set(ctx, td.items...)
			st.Require().NoError(setErr, "No error expected on setting value")

			for _, item := range td.items {
				st.Run("Loading key: "+item.Key, func() {
					dst := td.dst()
					getErr := st.cache.Get(ctx, &dst, item.Key)
					st.Require().NoError(getErr, "No error expected on getting value")

					st.Require().Equal(item.Value, dst, "Unexpected value returned from Redis")
				})
			}
		})
	}
}

func (st *CacheSuite) TestGet() {
	ctx := context.Background()

	st.Require().NoError(
		st.cache.SetKV(ctx, st.commonTestData.keyValPairs...),
		"No error expected on setting elements",
	)

	st.Run("get each single key one by one", func() {
		for k, v := range st.commonTestData.keyVals {
			var dst string
			st.Require().NoError(
				st.cache.Get(ctx, &dst, k),
				"No error expected on getting key "+k,
			)
			st.Require().Equal(v, dst, "Unexpected value for key "+k)
		}
	})

	st.Run("get non-exists key ignoring cache miss errors", func() {
		var dst string
		err := st.cache.Get(ctx, &dst, nonExistKey)
		st.Require().NoError(err, "No error expected on getting non-exist-key")
		st.Require().Empty(dst, "Dst should remain unchanged")
	})

	st.Run("get non-exists key with cache miss errors", func() {
		var dst string
		err := st.cache.AddCacheMissErrors().Get(ctx, &dst, nonExistKey)
		st.Require().Error(err, "An error expected")
		st.Require().Empty(dst, "Dst should remain unchanged")
	})

	st.Run("get all keys into a slice", func() {
		var dst []string
		st.Require().NoError(
			st.cache.Get(ctx, &dst, st.commonTestData.keys...),
			"Multi get failed",
		)
		st.Require().EqualValues(st.commonTestData.vals, dst)
	})

	st.Run("get single key into a slice", func() {
		var dst []string
		st.Require().NoError(
			st.cache.Get(ctx, &dst, st.commonTestData.keys[0]),
			"Multi get failed",
		)
		st.Require().EqualValues(st.commonTestData.vals[0:1], dst)
	})

	st.Run("get all keys into a map", func() {
		var dst map[string]string
		st.Require().NoError(
			st.cache.Get(ctx, &dst, st.commonTestData.keys...),
			"Multi get failed",
		)
		st.Require().EqualValues(st.commonTestData.keyVals, dst)
	})

	st.Run("get single key into a map", func() {
		var dst map[string]string
		st.Require().NoError(
			st.cache.Get(ctx, &dst, st.commonTestData.keys[0]),
			"Multi get failed",
		)
		st.Require().EqualValues(map[string]string{st.commonTestData.keys[0]: st.commonTestData.vals[0]}, dst)
	})
}

func (st *CacheSuite) TestGet_IntoContainer_WithoutMapKeyModification() {
	testDestinations := []struct {
		dst          func() interface{}
		expectedData func() interface{}
	}{
		{
			dst: func() interface{} {
				return map[string]string{}
			},
			expectedData: func() interface{} {
				return st.commonTestData.keyVals
			},
		},
		{
			dst: func() interface{} {
				return []string{}
			},
			expectedData: func() interface{} {
				return st.commonTestData.vals
			},
		},
	}

	testData := []struct {
		testCase string
		cache    *cache.Cache
		loadFn   func(absentKeys ...string) (interface{}, error)
	}{
		{
			testCase: "function returns kv-map",
			cache: st.cache.
				WithAbsentKeysLoader(func(absentKeys ...string) (interface{}, error) {
					m := map[string]string{}
					for _, k := range absentKeys {
						m[k] = st.commonTestData.keyVals[k]
					}
					return m, nil
				}),
		},
		{
			testCase: "function returns slice",
			cache: st.cache.
				WithAbsentKeysLoader(func(absentKeys ...string) (interface{}, error) {
					var s []string
					for _, k := range absentKeys {
						s = append(s, st.commonTestData.keyVals[k])
					}
					return s, nil
				}).
				WithItemToCacheKey(func(it interface{}) (string, string) {
					for k, v := range st.commonTestData.keyVals {
						if v == it.(string) {
							return k, ""
						}
					}
					return "", ""
				}),
		},
	}

	for _, dstData := range testDestinations {
		for _, td := range testData {
			st.SetupTest()
			dst := dstData.dst()
			st.Run(fmt.Sprintf("%s adds element into %T", td.testCase, dst), func() {
				getErr := td.cache.Get(
					context.Background(),
					&dst,
					st.commonTestData.keys...,
				)

				expectedData := dstData.expectedData()

				st.Require().NoError(getErr, "No error expected on loading keys from cache")
				checkDst(st.T(), expectedData, dst, "result data should be identical")
				st.checkKeysPresenceInCache()
			})
		}
	}
}

func (st *CacheSuite) TestGet_IntoMap_WithMapKeyModification() {
	dst := map[string]string{}

	expectedData := map[string]string{}
	for k, v := range st.commonTestData.keyVals {
		expectedData["converted_"+k] = v
	}

	getErr := st.cache.
		ConvertCacheKeyToMapKey(func(cacheKey string) string {
			return "converted_" + cacheKey
		}).
		WithAbsentKeysLoader(func(absentKeys ...string) (interface{}, error) {
			m := map[string]string{}
			for _, k := range absentKeys {
				m[k] = st.commonTestData.keyVals[k]
			}
			return m, nil
		}).
		Get(
			context.Background(),
			&dst,
			st.commonTestData.keys...,
		)

	st.Require().NoError(getErr, "No error expected on loading keys from cache")
	checkDst(st.T(), expectedData, dst, "result data should be identical")
	// check the data is cached by non-changed keys
	st.checkKeysPresenceInCache()
}

func (st *CacheSuite) checkKeysPresenceInCache() {
	st.T().Helper()
	for k, v := range st.commonTestData.keyVals {
		var dstSingle string
		singleKeyGetErr := st.cache.Get(context.Background(), &dstSingle, k)
		st.Require().NoErrorf(singleKeyGetErr, "No error expected on getting %q", k)
		st.Require().Equalf(v, dstSingle, "Unexpected value for key %q", k)
	}
}

func TestCacheSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &CacheSuite{})
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

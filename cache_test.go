package cache

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/suite"
	"syreclabs.com/go/faker"

	"github.com/vkuptcov/go-redis-cache/v8/internal/marshaller"
)

type CacheSuite struct {
	suite.Suite
	client         *redis.Client
	cache          *Cache
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

	st.cache = &Cache{opt: &Options{
		Redis:      st.client,
		DefaultTTL: 0,
		Marshaller: marshaller.NewMarshaller(&marshaller.JSONMarshaller{}),
	}}
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
		items    []*Item
		dst      func() interface{}
	}{
		{
			testCase: "set single item",
			items: []*Item{
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
			items: []*Item{
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
		err := st.cache.Get(WithCacheMissErrorsContext(ctx), &dst, nonExistKey)
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

func (st *CacheSuite) TestGetOrLoad() {
	dst := map[string]string{}
	getErr := st.cache.GetOrLoad(
		context.Background(),
		&dst,
		func(absentKeys ...string) (interface{}, error) {
			m := map[string]string{}
			for _, k := range absentKeys {
				m[k] = st.commonTestData.keyVals[k]
			}
			return m, nil
		},
		st.commonTestData.keys...,
	)

	expectedMap := st.commonTestData.keyVals

	st.Require().NoError(getErr, "No error expected on loading keys from cache")
	st.Require().Empty(cmp.Diff(expectedMap, dst), "", "maps should be identical")

	for k, v := range expectedMap {
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

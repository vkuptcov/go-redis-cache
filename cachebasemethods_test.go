package cache_test

//this file contains direct tests for base methods, such as Set, SetKV, HSetKV, HSetKV, HGetAll, HGetFieldsForKey and HGetKeysAndFields

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/suite"
	"syreclabs.com/go/faker"

	cache "github.com/vkuptcov/go-redis-cache/v8"
	"github.com/vkuptcov/go-redis-cache/v8/internal/marshaller"
)

type BaseMethodsSuite struct {
	suite.Suite
	client         *redis.Client
	cache          *cache.Cache
	commonTestData commonTestData
}

func (st *BaseMethodsSuite) SetupSuite() {
	st.client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	st.cache = cache.NewCache(cache.Options{
		Redis:      st.client,
		DefaultTTL: 0,
		Marshaller: marshaller.NewMarshaller(&marshaller.JSONMarshaller{}),
	})
}

func (st *BaseMethodsSuite) SetupTest() {
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

func (st *BaseMethodsSuite) TestSet() {
	ctx := context.Background()

	testData := []struct {
		testCase string
		items    []*cache.Item
	}{
		{
			testCase: "set single item",
			items: []*cache.Item{
				{
					Key:   faker.RandomString(10),
					Value: faker.Lorem().Sentence(5),
				},
			},
		},
		{
			testCase: "set single hash item",
			items: []*cache.Item{
				{
					Key:   faker.RandomString(10),
					Field: faker.RandomString(10),
					Value: faker.Lorem().Sentence(5),
				},
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
		},
		{
			testCase: "set several hash items",
			items: []*cache.Item{
				{
					Key:   faker.RandomString(10),
					Field: faker.RandomString(10),
					Value: faker.Lorem().Sentence(5),
				},
				{
					Key:   faker.RandomString(10),
					Field: faker.RandomString(10),
					Value: faker.Lorem().Sentence(5),
				},
				{
					Key:   faker.RandomString(10),
					Field: faker.RandomString(10),
					Value: faker.Lorem().Sentence(5),
				},
			},
		},
		{
			testCase: "set mixed hash and non-hash items",
			items: []*cache.Item{
				{
					Key:   faker.RandomString(10),
					Value: faker.Lorem().Sentence(5),
				},
				{
					Key:   faker.RandomString(10),
					Field: faker.RandomString(10),
					Value: faker.Lorem().Sentence(5),
				},
				{
					Key:   faker.RandomString(10),
					Field: faker.RandomString(10),
					Value: faker.Lorem().Sentence(5),
				},
				{
					Key:   faker.RandomString(10),
					Value: faker.Lorem().Sentence(5),
				},
			},
		},
	}
	for _, td := range testData {
		st.Run(td.testCase, func() {
			setErr := st.cache.Set(ctx, td.items...)
			st.Require().NoError(setErr, "No error expected on setting value")

			for _, item := range td.items {
				st.Run("Loading key: "+item.Key, func() {
					var stringCmd *redis.StringCmd
					if item.Field == "" {
						stringCmd = st.client.Get(ctx, item.Key)
					} else {
						stringCmd = st.client.HGet(ctx, item.Key, item.Field)
					}
					st.Require().NoError(stringCmd.Err(), "No error expected on getting value")

					st.Require().Equal(item.Value, stringCmd.Val(), "Unexpected value returned from Redis")
				})
			}
		})
	}
}

func (st *BaseMethodsSuite) TestGet() {
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

func TestBaseMethodsSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &BaseMethodsSuite{})
}

package cache

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/suite"
	"syreclabs.com/go/faker"
)

type CacheSuite struct {
	suite.Suite
	client *redis.Client
	cache  *Cache
}

func (st *CacheSuite) SetupSuite() {
	st.client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	st.cache = &Cache{opt: &Options{
		Redis:      st.client,
		DefaultTTL: 0,
		Marshaller: &baseMarshaller{
			customMarshaller: &testJSONMarshaller{},
		},
	}}
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
		{
			testCase: "set non-loaded item",
			items: []*Item{
				{
					Key: faker.RandomString(10),
					Load: func(item *Item) (interface{}, error) {
						return faker.Lorem().Sentence(5), nil
					},
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

	keyVals := map[string]interface{}{}
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
	st.Require().NoError(
		st.cache.SetKV(ctx, keyValPairs...),
		"No error expected on setting elements",
	)

	var dst []string
	st.Require().NoError(
		st.cache.Get(ctx, &dst, keys...),
		"Multi get failed",
	)
	st.Require().EqualValues(vals, dst)

	for k, v := range keyVals {
		var dst string
		st.Require().NoError(
			st.cache.Get(ctx, &dst, k),
			"No error expected on getting key "+k,
		)
		st.Require().Equal(v, dst, "Unexpected value for key "+k)
	}
}

func TestCacheSuite(t *testing.T) {
	suite.Run(t, &CacheSuite{})
}

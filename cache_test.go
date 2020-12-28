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

func TestCacheSuite(t *testing.T) {
	suite.Run(t, &CacheSuite{})
}

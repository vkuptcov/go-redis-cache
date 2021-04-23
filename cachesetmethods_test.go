package cache_test

import (
	"testing"

	"github.com/go-redis/redis/v7"
	"github.com/stretchr/testify/suite"
	"syreclabs.com/go/faker"

	cache "github.com/vkuptcov/go-redis-cache/v7"
)

type SetMethodsSuite struct {
	BaseCacheSuite
}

func (st *SetMethodsSuite) TestSet() {
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
					Field: faker.RandomString(8),
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
					Field: faker.RandomString(8),
					Value: faker.Lorem().Sentence(5),
				},
				{
					Key:   faker.RandomString(10),
					Field: faker.RandomString(8),
					Value: faker.Lorem().Sentence(5),
				},
				{
					Key:   faker.RandomString(10),
					Field: faker.RandomString(8),
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
					Field: faker.RandomString(8),
					Value: faker.Lorem().Sentence(5),
				},
				{
					Key:   faker.RandomString(10),
					Field: faker.RandomString(8),
					Value: faker.Lorem().Sentence(5),
				},
				{
					Key:   faker.RandomString(10),
					Value: faker.Lorem().Sentence(5),
				},
			},
		},
		{
			testCase: "set struct",
			items: []*cache.Item{
				{
					Key: faker.RandomString(10),
					Value: struct {
						A string
						B string
					}{
						A: faker.Lorem().Word(),
						B: faker.Lorem().Word(),
					},
				},
				{
					Key:   faker.RandomString(10),
					Field: faker.RandomString(8),
					Value: struct {
						D string
						C string
					}{
						D: faker.Lorem().Word(),
						C: faker.Lorem().Word(),
					},
				},
			},
		},
		{
			testCase: "set nil",
			items: []*cache.Item{
				{
					Key:   faker.RandomString(10),
					Value: nil,
				},
				{
					Key:   faker.RandomString(10),
					Field: faker.RandomString(8),
					Value: nil,
				},
			},
		},
	}
	for _, td := range testData {
		st.Run(td.testCase, func() {
			setErr := st.cache.Set(st.ctx, td.items...)
			st.Require().NoError(setErr, "No error expected on setting value")

			for _, item := range td.items {
				st.Run("Loading key: "+item.Key, func() {
					var stringCmd *redis.StringCmd
					if item.Field == "" {
						stringCmd = st.client.Get(item.Key)
					} else {
						stringCmd = st.client.HGet(item.Key, item.Field)
					}
					expected, marshalErr := st.marshaller.Marshal(item.Value)
					st.Require().NoError(marshalErr, "")
					st.Require().NoError(stringCmd.Err(), "No error expected on getting value")

					st.Require().Equal(string(expected), stringCmd.Val(), "Unexpected value returned from Redis")
				})
			}
		})
	}
}

func (st *SetMethodsSuite) TestSetKV() {
	keyVals := st.generateKeyValPairs()

	setErr := st.cache.SetKV(st.ctx, keyVals.keyValPairs...)
	st.Require().NoError(setErr, "No error expected for SetKV")
	for k, v := range keyVals.keyVals {
		stringCmd := st.client.Get(k)
		st.Require().NoError(stringCmd.Err(), "No error expected on getting value")
		st.Require().EqualValues(v, stringCmd.Val(), "Unexpected value")
	}
}

func (st *SetMethodsSuite) TestHSetKV() {
	key := faker.RandomString(7)
	fieldValPairs := st.generateKeyValPairs()

	setErr := st.cache.HSetKV(st.ctx, key, fieldValPairs.keyValPairs...)
	st.Require().NoError(setErr, "No error expected for SetKV")
	for f, v := range fieldValPairs.keyVals {
		stringCmd := st.client.HGet(key, f)
		st.Require().NoError(stringCmd.Err(), "No error expected on getting value")
		st.Require().EqualValues(v, stringCmd.Val(), "Unexpected value")
	}
}

func TestSetMethodsSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &SetMethodsSuite{})
}

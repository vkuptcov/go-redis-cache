package cache_test

import (
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/suite"
	"syreclabs.com/go/faker"

	cache "github.com/vkuptcov/go-redis-cache/v8"
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
						stringCmd = st.client.Get(st.ctx, item.Key)
					} else {
						stringCmd = st.client.HGet(st.ctx, item.Key, item.Field)
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
	var keyVals []interface{}
	for i := 0; i < 3; i++ {
		keyVals = append(keyVals, faker.RandomString(5), faker.Lorem().Word())
	}
	setErr := st.cache.SetKV(st.ctx, keyVals...)
	st.Require().NoError(setErr, "No error expected for SetKV")
	for i := 0; i < len(keyVals); i += 2 {
		stringCmd := st.client.Get(st.ctx, keyVals[i].(string))
		st.Require().NoError(stringCmd.Err(), "No error expected on getting value")
		st.Require().EqualValues(keyVals[i+1], stringCmd.Val(), "Unexpected value")
	}
}

func (st *SetMethodsSuite) TestHSetKV() {
	key := faker.RandomString(7)
	var fieldValPairs []interface{}
	for i := 0; i < 3; i++ {
		fieldValPairs = append(fieldValPairs, faker.RandomString(5), faker.Lorem().Word())
	}
	setErr := st.cache.HSetKV(st.ctx, key, fieldValPairs...)
	st.Require().NoError(setErr, "No error expected for SetKV")
	for i := 0; i < len(fieldValPairs); i += 2 {
		stringCmd := st.client.HGet(st.ctx, key, fieldValPairs[i].(string))
		st.Require().NoError(stringCmd.Err(), "No error expected on getting value")
		st.Require().EqualValues(fieldValPairs[i+1], stringCmd.Val(), "Unexpected value")
	}
}

func TestBaseMethodsSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &SetMethodsSuite{})
}

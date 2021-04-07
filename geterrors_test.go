package cache_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"syreclabs.com/go/faker"

	cache "github.com/vkuptcov/go-redis-cache/v8"
)

type GetErrorsSuite struct {
	BaseCacheSuite
	scalarKey string
	scalarVal string
	hashKey   string
	hashField string
	hashVal   string
}

func (st *GetErrorsSuite) SetupTest() {
	st.scalarKey = faker.RandomString(5)
	st.scalarVal = faker.RandomString(7)
	st.hashKey = faker.RandomString(5)
	st.hashField = faker.RandomString(3)
	st.hashVal = faker.RandomString(7)

	st.Require().NoError(
		st.cache.Set(
			st.ctx,
			&cache.Item{
				Key:   st.scalarKey,
				Value: st.scalarVal,
			},
			&cache.Item{
				Key:   st.hashKey,
				Field: st.hashField,
				Value: st.hashVal,
			},
		),
		"No error expected on preparing test data",
	)
}

func (st *GetErrorsSuite) TestKeyMissErrorsForDefaultSettings() {
	testCases := []struct {
		testCase string
		loader   func(dst interface{}) error
	}{
		{
			testCase: "load scalars (one of key exists)",
			loader: func(dst interface{}) error {
				return st.cache.Get(st.ctx, dst, st.scalarKey, faker.RandomString(15))
			},
		},
		{
			testCase: "load scalars with a single non-exists key",
			loader: func(dst interface{}) error {
				return st.cache.Get(st.ctx, &dst, faker.RandomString(15))
			},
		},
	}

	for _, tc := range testCases {
		st.Run(tc.testCase, func() {
			st.Run("load into slice", func() {
				var dst []string
				st.Require().NoError(
					tc.loader(&dst),
					"No error expected on loading elements",
				)
			})
			st.Run("load into map", func() {
				dst := map[string]string{}
				st.Require().NoError(
					tc.loader(&dst),
					"No error expected on loading elements",
				)
			})
		})
	}
}

func TestGetErrorsSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &GetErrorsSuite{})
}

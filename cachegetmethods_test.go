package cache_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"syreclabs.com/go/faker"
)

type GetMethodsSuite struct {
	BaseCacheSuite
}

func (st *GetMethodsSuite) TestGet() {
	keyVals := st.generateKeyValPairs()
	st.Require().NoError(
		st.cache.SetKV(st.ctx, keyVals.keyValPairs...),
		"No error expected on setting element",
	)
	st.Run("load into a primitive", func() {
		var dst string
		st.Require().NoError(
			st.cache.Get(st.ctx, &dst, keyVals.keys[0]),
			"No error expected on getting key",
		)
		st.Require().EqualValues(keyVals.vals[0], dst, "Unexpected dst")
	})
	st.Run("load into a slice", func() {
		var dst []string
		st.Require().NoError(
			st.cache.Get(st.ctx, &dst, keyVals.keys...),
			"No error expected on getting key",
		)
		checkDst(st.T(), keyVals.vals, dst, "Unexpected dst")
	})
	st.Run("load into a map", func() {
		var dst map[string]string
		st.Require().NoError(
			st.cache.Get(st.ctx, &dst, keyVals.keys...),
			"No error expected on getting key",
		)
		checkDst(st.T(), keyVals.keyVals, dst, "Unexpected dst")
	})
}

func (st *GetMethodsSuite) TestHGetAll() {
	ctx := context.Background()

	firstKey := faker.RandomString(7)
	firstKeyData := st.generateKeyValPairs()
	secondKey := faker.RandomString(7)
	secondKeyData := st.generateKeyValPairs()

	st.Require().NoError(st.cache.HSetKV(ctx, firstKey, firstKeyData.keyValPairs...), "first key values set err")
	st.Require().NoError(st.cache.HSetKV(ctx, secondKey, secondKeyData.keyValPairs...), "second key values set err")

	st.Run("load into a slice", func() {
		var dst []string
		st.Require().NoError(
			st.cache.HGetAll(st.ctx, &dst, firstKey, secondKey),
			"No error expected on getting key",
		)
		expectedData := append([]string{}, firstKeyData.vals...)
		expectedData = append(expectedData, secondKeyData.vals...)
		checkDst(st.T(), expectedData, dst, "Unexpected dst")
	})
	st.Run("load into a map", func() {
		var dst map[string]string
		st.Require().NoError(
			st.cache.HGetAll(st.ctx, &dst, firstKey, secondKey),
			"No error expected on getting key",
		)
		expectedData := map[string]string{}
		addIntoMap := func(key string, fieldVals map[string]string) {
			for f, v := range fieldVals {
				expectedData[key+"-"+f] = v
			}
		}
		addIntoMap(firstKey, firstKeyData.keyVals)
		addIntoMap(secondKey, secondKeyData.keyVals)
		checkDst(st.T(), expectedData, dst, "Unexpected dst")
	})
}

func TestMethodsSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &GetMethodsSuite{})
}

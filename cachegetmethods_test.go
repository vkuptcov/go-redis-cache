package cache_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"syreclabs.com/go/faker"

	"github.com/vkuptcov/go-redis-cache/v8/cachekeys"
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
	st.Run("load into a primitive set as interface", func() {
		var dst interface{} = ""
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
	hashMapData := st.prepareHashMapData()
	var keys []string
	for k := range hashMapData {
		keys = append(keys, k)
	}

	st.Run("load into a slice", func() {
		var dst []string
		st.Require().NoError(
			st.cache.HGetAll(st.ctx, &dst, keys...),
			"No error expected on getting keysToLoad",
		)

		var expectedData []string
		for _, d := range hashMapData {
			expectedData = append(expectedData, d.vals...)
		}
		checkDst(st.T(), expectedData, dst, "Unexpected dst")
	})
	st.Run("load into a map", func() {
		var dst map[string]string
		st.Require().NoError(
			st.cache.HGetAll(st.ctx, &dst, keys...),
			"No error expected on getting keysToLoad",
		)
		expectedData := map[string]string{}
		for k, d := range hashMapData {
			for f, v := range d.keyVals {
				expectedData[cachekeys.KeyWithField(k, f)] = v
			}
		}
		checkDst(st.T(), expectedData, dst, "Unexpected dst")
	})
	st.Run("load into a map of maps", func() {
		var dst map[string]map[string]string
		st.Require().NoError(
			st.cache.HGetAll(st.ctx, &dst, keys...),
			"No error expected on getting keysToLoad",
		)
		expectedData := map[string]map[string]string{}
		for k, d := range hashMapData {
			for f, v := range d.keyVals {
				if _, exists := expectedData[k]; !exists {
					expectedData[k] = map[string]string{}
				}
				expectedData[k][f] = v
			}
		}
		checkDst(st.T(), expectedData, dst, "Unexpected dst")
	})
}

func (st *GetMethodsSuite) TestHGetKeysAndFields() {
	hashMapData := st.prepareHashMapData()
	keysToFields := map[string][]string{}
	for k, d := range hashMapData {
		keysToFields[k] = d.keys[0:2]
	}

	st.Run("load into a slice", func() {
		var dst []string
		st.Require().NoError(
			st.cache.HGetKeysAndFields(st.ctx, &dst, keysToFields),
			"No error expected on getting keysToLoad",
		)
		var expectedData []string
		for _, d := range hashMapData {
			expectedData = append(expectedData, d.vals[0:2]...)
		}
		checkDst(st.T(), expectedData, dst, "Unexpected dst")
	})
	st.Run("load into a map", func() {
		var dst map[string]string
		st.Require().NoError(
			st.cache.HGetKeysAndFields(st.ctx, &dst, keysToFields),
			"No error expected on getting keysToLoad",
		)
		expectedData := map[string]string{}
		for k, d := range hashMapData {
			for idx, f := range d.keys[0:2] {
				expectedData[cachekeys.KeyWithField(k, f)] = d.vals[idx]
			}
		}
		checkDst(st.T(), expectedData, dst, "Unexpected dst")
	})
}

func (st *GetMethodsSuite) TestHGetFieldsForKey() {
	hashMapData := st.prepareHashMapData()
	st.Run("load into a slice", func() {
		var dst []string
		for k, d := range hashMapData {
			st.Require().NoError(
				st.cache.HGetFieldsForKey(st.ctx, &dst, k, d.keys...),
				"No error expected on getting keysToLoad",
			)
			checkDst(st.T(), d.vals, dst, "Unexpected dst")
			break
		}
	})
	st.Run("load into a map", func() {
		var dst map[string]string
		for k, d := range hashMapData {
			st.Require().NoError(
				st.cache.HGetFieldsForKey(st.ctx, &dst, k, d.keys...),
				"No error expected on getting keysToLoad",
			)
			expectedData := map[string]string{}
			for idx, f := range d.keys {
				expectedData[cachekeys.KeyWithField(k, f)] = d.vals[idx]
			}

			checkDst(st.T(), expectedData, dst, "Unexpected dst")
			break
		}
	})
}

func (st *GetMethodsSuite) TestLoadEmptyKeysForGetMethods() {
	st.Run("empty keys for Get", func() {
		var dst map[string]string
		st.Require().NoError(
			st.cache.Get(st.ctx, &dst),
			"no error expected",
		)
	})
	st.Run("empty keys for HGetAll", func() {
		var dst map[string]string
		st.Require().NoError(
			st.cache.HGetAll(st.ctx, &dst),
			"no error expected",
		)
	})
	st.Run("empty fields for HGetFieldsForKey", func() {
		var dst map[string]string
		st.Require().NoError(
			st.cache.HGetFieldsForKey(st.ctx, &dst, "some-key"),
			"no error expected",
		)
	})
	st.Run("empty map for HGetKeysAndFields", func() {
		var dst map[string]string
		st.Require().NoError(
			st.cache.HGetKeysAndFields(st.ctx, &dst, map[string][]string{}),
			"no error expected",
		)
	})
	st.Run("map with empty fields for HGetKeysAndFields", func() {
		var dst map[string]string
		st.Require().NoError(
			st.cache.HGetKeysAndFields(st.ctx, &dst, map[string][]string{
				"some-key": {},
			}),
			"no error expected",
		)
	})
}

func (st *GetMethodsSuite) prepareHashMapData() map[string]commonTestData {
	st.T().Helper()
	firstKey := faker.RandomString(7)
	firstKeyData := st.generateKeyValPairs()
	secondKey := faker.RandomString(7)
	secondKeyData := st.generateKeyValPairs()

	st.Require().NoError(st.cache.HSetKV(st.ctx, firstKey, firstKeyData.keyValPairs...), "first key values set err")
	st.Require().NoError(st.cache.HSetKV(st.ctx, secondKey, secondKeyData.keyValPairs...), "second key values set err")

	return map[string]commonTestData{
		firstKey:  firstKeyData,
		secondKey: secondKeyData,
	}
}

func TestMethodsSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &GetMethodsSuite{})
}

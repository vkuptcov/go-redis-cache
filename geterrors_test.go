package cache_test

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"github.com/vkuptcov/go-redis-cache/v7/cachekeys"
	"syreclabs.com/go/faker"

	cache "github.com/vkuptcov/go-redis-cache/v7"
)

type GetErrorsSuite struct {
	BaseCacheSuite
	scalarKey string
	scalarVal string
	hashKey   string
	hashField string
	hashVal   string
}

const nonExistedKey = "non-existed-key"
const nonExistedField = "non-existed-field"

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

func (st *GetErrorsSuite) TestKeyMissErrors() {
	testCases := []struct {
		testCase       string
		loader         func(c *cache.Cache, dst interface{}) error
		cacheMissError *cache.KeyErr
	}{
		{
			testCase: "load scalars (one of key exists)",
			loader: func(c *cache.Cache, dst interface{}) error {
				return c.Get(st.ctx, dst, st.scalarKey, nonExistedKey)
			},
			cacheMissError: &cache.KeyErr{
				KeysToErrs: map[string]error{
					nonExistedKey: cache.ErrCacheMiss,
				},
			},
		},
		{
			testCase: "load scalars with a single non-exists key",
			loader: func(c *cache.Cache, dst interface{}) error {
				return c.Get(st.ctx, dst, nonExistedKey)
			},
			cacheMissError: &cache.KeyErr{
				KeysToErrs: map[string]error{
					nonExistedKey: cache.ErrCacheMiss,
				},
			},
		},
		{
			testCase: "load hash map (only one of key exists)",
			loader: func(c *cache.Cache, dst interface{}) error {
				return c.HGetAll(st.ctx, dst, st.hashKey, nonExistedKey)
			},
			cacheMissError: &cache.KeyErr{
				KeysToErrs: map[string]error{
					nonExistedKey: cache.ErrCacheMiss,
				},
			},
		},
		{
			testCase: "load hash map with a single non-exists key",
			loader: func(c *cache.Cache, dst interface{}) error {
				return c.HGetAll(st.ctx, dst, nonExistedKey)
			},
			cacheMissError: &cache.KeyErr{
				KeysToErrs: map[string]error{
					nonExistedKey: cache.ErrCacheMiss,
				},
			},
		},
		{
			testCase: "load hash map and fields for existing key and one of existing fields",
			loader: func(c *cache.Cache, dst interface{}) error {
				return c.HGetFieldsForKey(st.ctx, dst, st.hashKey, st.hashField, nonExistedField)
			},
			cacheMissError: &cache.KeyErr{
				KeysToErrs: map[string]error{
					cachekeys.KeyWithField(st.hashKey, nonExistedField): cache.ErrCacheMiss,
				},
			},
		},
		{
			testCase: "load hash map and fields for existing key and one of non-existing fields",
			loader: func(c *cache.Cache, dst interface{}) error {
				return c.HGetFieldsForKey(st.ctx, dst, st.hashKey, nonExistedField)
			},
			cacheMissError: &cache.KeyErr{
				KeysToErrs: map[string]error{
					cachekeys.KeyWithField(st.hashKey, nonExistedField): cache.ErrCacheMiss,
				},
			},
		},
		{
			testCase: "load hash map and fields for non-existing key",
			loader: func(c *cache.Cache, dst interface{}) error {
				return c.HGetFieldsForKey(st.ctx, dst, nonExistedKey, nonExistedField)
			},
			cacheMissError: &cache.KeyErr{
				KeysToErrs: map[string]error{
					cachekeys.KeyWithField(nonExistedKey, nonExistedField): cache.ErrCacheMiss,
				},
			},
		},
		{
			testCase: "load hash map and fields for non-existing keys and fields",
			loader: func(c *cache.Cache, dst interface{}) error {
				return c.HGetKeysAndFields(st.ctx, dst, map[string][]string{
					st.hashKey:    {st.hashField, nonExistedField},
					nonExistedKey: {nonExistedField},
				})
			},
			cacheMissError: &cache.KeyErr{
				KeysToErrs: map[string]error{
					cachekeys.KeyWithField(st.hashKey, nonExistedField):    cache.ErrCacheMiss,
					cachekeys.KeyWithField(nonExistedKey, nonExistedField): cache.ErrCacheMiss,
				},
			},
		},
	}

	st.Run("DefaultSettings", func() {
		for _, tc := range testCases {
			st.Run(tc.testCase, func() {
				st.Run("load into slice", func() {
					var dst []string
					st.Require().NoError(
						tc.loader(st.cache, &dst),
						"No error expected on loading elements",
					)
				})
				st.Run("load into map", func() {
					dst := map[string]string{}
					st.Require().NoError(
						tc.loader(st.cache, &dst),
						"No error expected on loading elements",
					)
				})
			})
		}
	})

	st.Run("AddCacheMissError", func() {
		c := st.cache.AddCacheMissErrors()
		for _, tc := range testCases {
			st.Run(tc.testCase, func() {
				st.Run("load into slice", func() {
					var dst []string
					loadErr := tc.loader(c, &dst)
					st.compareKeyErr(tc.cacheMissError, loadErr)
				})
				st.Run("load into map", func() {
					dst := map[string]string{}
					loadErr := tc.loader(c, &dst)
					st.compareKeyErr(tc.cacheMissError, loadErr)
				})
			})
		}
	})
}

func (st *GetErrorsSuite) Test_SingleElementDestination_ShouldHaveCacheMissErrorByDefault() {
	var dst string
	testCases := []struct {
		testCase string
		loader   func() error
	}{
		{
			testCase: "simple get",
			loader: func() error {
				return st.cache.Get(st.ctx, &dst, nonExistedKey)
			},
		},
		{
			testCase: "hget fields for single key",
			loader: func() error {
				return st.cache.HGetFieldsForKey(st.ctx, &dst, nonExistedKey, nonExistedField)
			},
		},
		{
			testCase: "hget all",
			loader: func() error {
				return st.cache.HGetAll(st.ctx, &dst, nonExistedKey)
			},
		},
		{
			testCase: "hget keys and fields",
			loader: func() error {
				return st.cache.HGetKeysAndFields(st.ctx, &dst, map[string][]string{
					nonExistedKey: {nonExistedField},
				})
			},
		},
	}
	for _, tc := range testCases {
		st.Run(tc.testCase, func() {
			dst = ""
			loadErr := tc.loader()
			st.Require().Truef(errors.Is(loadErr, cache.ErrCacheMiss), "cache.ErrCacheMiss expected, %+v given", loadErr)
		})
	}
}

func (st *GetErrorsSuite) compareKeyErr(expected *cache.KeyErr, actual error) {
	st.T().Helper()
	st.Require().Errorf(actual, "Error expected")
	var keyErr *cache.KeyErr
	st.Require().Truef(errors.As(actual, &keyErr), "*cache.KeyErr expected, %T given", actual)

	st.Require().Equal(len(expected.KeysToErrs), len(keyErr.KeysToErrs), "Unexpected number of keys in error")
	st.Require().Equal(len(keyErr.KeysToErrs), keyErr.CacheMissErrsCount, "Unexpected cache miss err count")

	for k, err := range expected.KeysToErrs {
		if kErr, exists := keyErr.KeysToErrs[k]; exists {
			st.Require().Truef(errors.Is(kErr, err), "%T:%+v error expected, %T:%+v given", err, err, kErr, kErr)
		} else {
			st.Failf("Expected an error", " for key %q", k)
		}
	}
}

func TestGetErrorsSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &GetErrorsSuite{})
}

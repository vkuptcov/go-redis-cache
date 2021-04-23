package cache_test

import (
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"syreclabs.com/go/faker"

	cache "github.com/vkuptcov/go-redis-cache/v8"
	"github.com/vkuptcov/go-redis-cache/v8/cachekeys"
)

type CacheAbsentKeysLoaderSuite struct {
	BaseCacheSuite
}

func (st *CacheAbsentKeysLoaderSuite) TestViaGet() {
	testCases := []struct {
		testCase        string
		keysToLoad      []string
		loader          func(t *testing.T, absentKeys ...string) interface{}
		expected        func(keys ...string) map[string]string
		expectedInCache func(keys ...string) map[string]string
	}{
		{
			testCase:   "returns single element",
			keysToLoad: []string{faker.RandomString(5)},
			loader: func(t *testing.T, absentKeys ...string) interface{} {
				t.Helper()
				require.New(t).Len(absentKeys, 1, "Only one key should be passed here")
				return st.keyToElement(absentKeys[0])
			},
			expected: st.keysToMap,
		},
		{
			testCase:   "returns single element as an Item",
			keysToLoad: []string{faker.RandomString(5)},
			loader: func(t *testing.T, absentKeys ...string) interface{} {
				t.Helper()
				require.New(t).Len(absentKeys, 1, "Only one key should be passed here")
				return &cache.Item{
					Key:   absentKeys[0],
					Value: st.keyToElement(absentKeys[0]),
				}
			},
			expected: st.keysToMap,
		},
		{
			testCase:   "returns a map",
			keysToLoad: []string{faker.RandomString(5), faker.RandomString(5), faker.RandomString(5)},
			loader: func(_ *testing.T, absentKeys ...string) interface{} {
				return st.keysToMap(absentKeys...)
			},
			expected: st.keysToMap,
		},
		{
			testCase:   "returns a mixed map with items and real values",
			keysToLoad: []string{faker.RandomString(5), faker.RandomString(5), faker.RandomString(5)},
			loader: func(t *testing.T, absentKeys ...string) interface{} {
				t.Helper()
				require.New(t).GreaterOrEqual(len(absentKeys), 2, "More than 1 key should be presented in the test")
				m := map[string]interface{}{}
				for idx, k := range absentKeys {
					if idx%2 == 0 {
						m[k] = st.keyToElement(k)
					} else {
						m[k] = &cache.Item{
							Key:   k,
							Value: st.keyToElement(k),
						}
					}
				}
				return m
			},
			expected: st.keysToMap,
		},
		{
			testCase:   "returns a slice of items",
			keysToLoad: []string{faker.RandomString(5), faker.RandomString(5), faker.RandomString(5)},
			loader: func(_ *testing.T, absentKeys ...string) interface{} {
				m := map[string]interface{}{}
				for _, k := range absentKeys {
					m[k] = &cache.Item{
						Key:   k,
						Value: st.keyToElement(k),
					}
				}
				return m
			},
			expected: st.keysToMap,
		},
	}
	for _, tc := range testCases {
		st.Run(tc.testCase, func() {
			var dst map[string]string
			st.Require().NoError(
				st.cache.WithAbsentKeysLoader(func(absentKeys ...string) (interface{}, error) {
					return tc.loader(st.T(), absentKeys...), nil
				}).Get(
					st.ctx,
					&dst,
					tc.keysToLoad...,
				),
				"No error expected for loading with absent keysToLoad",
			)

			expected := tc.expected(tc.keysToLoad...)
			st.Require().EqualValues(expected, dst, "unexpected loaded item")
			if tc.expectedInCache != nil {
				expected = tc.expectedInCache(tc.keysToLoad...)
			}
			st.checkElementsInCache(expected)
		})
	}
}

func (st *CacheAbsentKeysLoaderSuite) TestViaGet_ReturningSlice() {
	keysToLoad := []string{faker.RandomString(5), faker.RandomString(5), faker.RandomString(5)}
	var dst map[string]string
	st.Require().NoError(
		st.cache.
			ExtractCacheKeyWith(func(it interface{}) (key, field string) {
				return strings.TrimSuffix(it.(string), "-element"), ""
			}).
			WithAbsentKeysLoader(func(absentKeys ...string) (interface{}, error) {
				var s []string
				for _, k := range absentKeys {
					s = append(s, st.keyToElement(k))
				}
				return s, nil
			}).
			Get(
				st.ctx,
				&dst,
				keysToLoad...,
			),
		"No error expected for loading with absent keysToLoad",
	)
	expected := st.keysToMap(keysToLoad...)
	st.Require().EqualValues(expected, dst, "unexpected loaded item")
}

func (st *CacheAbsentKeysLoaderSuite) TestViaGet_FailsWithoutWithItemToCacheKeyFn() {
	var dst map[string]string
	cacheErr := st.cache.
		WithAbsentKeysLoader(func(absentKeys ...string) (interface{}, error) {
			var s []string
			for _, k := range absentKeys {
				s = append(s, st.keyToElement(k))
			}
			return s, nil
		}).
		Get(
			st.ctx,
			&dst,
			faker.RandomString(5), faker.RandomString(5), faker.RandomString(5),
		)
	st.Require().Truef(errors.Is(cacheErr, cache.ErrItemToCacheKeyFnRequired), "ErrItemToCacheKeyFnRequired expected, %+v of type %T given", cacheErr, cacheErr)
}

func (st *CacheAbsentKeysLoaderSuite) TestViaHGetFieldsForKey() {
	testCases := []struct {
		testCase     string
		key          string
		fields       []string
		absentLoader func(t *testing.T, absentKeys ...string) interface{}
	}{
		{
			testCase: "returns a single element",
			key:      faker.RandomString(5),
			fields: []string{
				faker.RandomString(7),
			},
			absentLoader: func(t *testing.T, absentKeys ...string) interface{} {
				t.Helper()
				require.New(t).Len(absentKeys, 1, "1 key expected")
				return st.keyToElement(absentKeys[0])
			},
		},
		{
			testCase: "returns a single Item",
			key:      faker.RandomString(5),
			fields: []string{
				faker.RandomString(7),
			},
			absentLoader: func(t *testing.T, absentKeys ...string) interface{} {
				t.Helper()
				require.New(t).Len(absentKeys, 1, "1 key expected")
				var k, f string
				cachekeys.UnpackKeyWithPrefix(absentKeys[0], &k, &f)
				return &cache.Item{
					Key:   k,
					Field: f,
					Value: st.keyToElement(absentKeys[0]),
				}
			},
		},
		{
			testCase: "returns a slice of items",
			key:      faker.RandomString(5),
			fields: []string{
				faker.RandomString(7),
				faker.RandomString(7),
			},
			absentLoader: func(t *testing.T, absentKeys ...string) interface{} {
				t.Helper()
				require.New(t).Len(absentKeys, 2, "2 keys expected")
				var items []*cache.Item
				for _, ak := range absentKeys {
					var k, f string
					cachekeys.UnpackKeyWithPrefix(ak, &k, &f)
					items = append(items, &cache.Item{
						Key:   k,
						Field: f,
						Value: st.keyToElement(ak),
					})
				}
				return items
			},
		},
		{
			testCase: "returns a map",
			key:      faker.RandomString(5),
			fields: []string{
				faker.RandomString(7),
				faker.RandomString(7),
			},
			absentLoader: func(t *testing.T, absentKeys ...string) interface{} {
				t.Helper()
				require.New(t).Len(absentKeys, 2, "2 keys expected")
				m := map[string]string{}
				for _, ak := range absentKeys {
					var k, f string
					cachekeys.UnpackKeyWithPrefix(ak, &k, &f)
					m[ak] = st.keyToElement(ak)
				}
				return m
			},
		},
	}
	for _, tc := range testCases {
		st.Run(tc.testCase, func() {
			var dst map[string]string
			st.Require().NoError(
				st.cache.
					WithAbsentKeysLoader(func(absentKeys ...string) (interface{}, error) {
						return tc.absentLoader(st.T(), absentKeys...), nil
					}).
					HGetFieldsForKey(
						st.ctx,
						&dst,
						tc.key,
						tc.fields...,
					),
				"No error expected",
			)
			expected := map[string]string{}
			for _, f := range tc.fields {
				k := cachekeys.KeyWithField(tc.key, f)
				expected[k] = st.keyToElement(k)
			}
			checkDst(st.T(), expected, dst, "Unmatched dst")

			var cachedDst map[string]string
			st.Require().NoError(
				st.cache.
					HGetFieldsForKey(
						st.ctx,
						&cachedDst,
						tc.key,
						tc.fields...,
					),
				"No error expected on getting cached object",
			)
			checkDst(st.T(), expected, cachedDst, "Unmatched cachedDst")
		})
	}
}

func TestCacheAbsentKeysLoaderSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &CacheAbsentKeysLoaderSuite{})
}

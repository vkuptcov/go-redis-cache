package cache_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"syreclabs.com/go/faker"
)

type CacheAbsentKeysLoaderSuite struct {
	BaseCacheSuite
}

func (st *CacheAbsentKeysLoaderSuite) TestReturningSingleElement() {
	keyToLoad := faker.RandomString(5)

	keyToElement := func(k string) string {
		return k + "-element"
	}

	testCases := []struct {
		testCase string
		dst      interface{}
		expected interface{}
	}{
		{
			testCase: "load into a string",
			dst:      "",
			expected: keyToElement(keyToLoad),
		},
		{
			testCase: "load into a slice",
			dst:      []string{},
			expected: []string{keyToElement(keyToLoad)},
		},
		{
			testCase: "load into a map",
			dst:      map[string]string{},
			expected: map[string]string{keyToLoad: keyToElement(keyToLoad)},
		},
	}
	for _, tc := range testCases {
		st.Run(tc.testCase, func() {
			dst := tc.dst
			st.Require().NoError(
				st.cache.WithAbsentKeysLoader(func(absentKeys ...string) (interface{}, error) {
					st.Require().Len(absentKeys, 1, "Only one key should be passed here")
					return keyToElement(absentKeys[0]), nil
				}).Get(
					st.ctx,
					&dst,
					keyToLoad,
				),
				"No error expected for loading with absent keys",
			)

			st.Require().EqualValues(tc.expected, dst, "unexpected loaded item")
		})
	}
}

func TestCacheAbsentKeysLoaderSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &CacheAbsentKeysLoaderSuite{})
}

package cache_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"syreclabs.com/go/faker"
)

type CacheWithItemToCacheKeySuite struct {
	BaseCacheSuite
}

func (st *CacheWithItemToCacheKeySuite) TestMapKeyTransformation() {
	var dst map[string]string
	key := faker.RandomString(5)
	transformKey := func(str string) string {
		return "transformed-key-for-" + str
	}
	st.Require().NoError(
		st.cache.
			WithItemToCacheKey(func(it interface{}) (key, field string) {
				return transformKey(it.(string)), ""
			}).
			WithAbsentKeysLoader(func(absentKeys ...string) (interface{}, error) {
				return st.keysToMap(absentKeys...), nil
			}).Get(
			context.Background(),
			&dst,
			key,
		),
		"No error expected on loading from cache",
	)

	element := st.keyToElement(key)
	st.checkElementsInCache(map[string]string{
		transformKey(element): element,
	})
}

func TestCacheWithItemToCacheKeySuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &CacheWithItemToCacheKeySuite{})
}

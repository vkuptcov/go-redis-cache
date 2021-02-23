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

	var dst string
	st.Require().NoError(
		st.cache.WithAbsentKeysLoader(func(absentKeys ...string) (interface{}, error) {
			st.Require().Len(absentKeys, 1, "Only one key should be passed here")
			return absentKeys[0] + "-element", nil
		}).Get(
			st.ctx,
			&dst,
			keyToLoad,
		),
		"No error expected for loading with absent keys",
	)

	st.Require().Equal(keyToLoad+"-element", dst, "unexpected loaded item")
}

func TestCacheAbsentKeysLoaderSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &CacheAbsentKeysLoaderSuite{})
}

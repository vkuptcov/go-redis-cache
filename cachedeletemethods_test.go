package cache_test

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	cache "github.com/vkuptcov/go-redis-cache/v8"
	"syreclabs.com/go/faker"
)

type DeleteMethodsSuite struct {
	BaseCacheSuite
}

func (st *DeleteMethodsSuite) TestDeleteSingleElement() {
	k := faker.RandomString(5)
	v := faker.RandomString(10)
	st.Require().NoErrorf(
		st.cache.SetKV(st.ctx, k, v),
		"No error expected on setting %q",
		k,
	)

	st.Require().NoError(
		st.cache.Delete(st.ctx, k),
		"No error expected on delete",
	)

	getErr := st.cache.Get(st.ctx, nil, k)
	st.Require().Truef(errors.Is(getErr, cache.ErrCacheMiss), "%q error expected, %q given", cache.ErrCacheMiss, getErr)
}

func (st *DeleteMethodsSuite) TestDeleteMultipleElements() {
	var keys []string
	for i := 0; i < 5; i++ {
		k := faker.RandomString(5)
		v := faker.RandomInt(10, 100)
		st.Require().NoErrorf(
			st.cache.SetKV(st.ctx, k, v),
			"No error expected on setting %q",
			k,
		)
		keys = append(keys, k)
	}

	st.Require().NoError(
		st.cache.Delete(st.ctx, keys...),
		"No error expected on delete",
	)

	var dst map[string]int
	st.Require().NoError(
		st.cache.Get(st.ctx, &dst, keys...),
		"No error expected on getting elements",
	)
	st.Require().Empty(dst, "No elements should be loaded after deletion")
}

func (st *DeleteMethodsSuite) TestDeleteEmptyKeysSlice() {
	var keys []string
	st.Require().NoError(
		st.cache.Delete(st.ctx, keys...),
		"No error expected on delete",
	)
}

func (st *DeleteMethodsSuite) TestDeleteNonExistsKeys() {
	st.Run("one key", func() {
		st.Require().NoError(
			st.cache.Delete(st.ctx, "non-exists-key"+faker.RandomString(10)),
			"No error expected on delete",
		)
	})
	st.Run("several keys", func() {
		st.Require().NoError(
			st.cache.Delete(
				st.ctx,
				"non-exists-key"+faker.RandomString(10),
				"non-exists-key"+faker.RandomString(10),
			),
			"No error expected on delete",
		)
	})
}

func TestDeleteMethodsSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &DeleteMethodsSuite{})
}

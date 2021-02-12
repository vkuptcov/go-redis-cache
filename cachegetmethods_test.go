package cache_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
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

func TestMethodsSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &GetMethodsSuite{})
}

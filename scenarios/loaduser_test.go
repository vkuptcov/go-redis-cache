package scenarios

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/suite"

	cache "github.com/vkuptcov/go-redis-cache/v7"
	"github.com/vkuptcov/go-redis-cache/v7/cachekeys"
)

type LoadUserSuite struct {
	BaseCacheSuite
	users           []*User
	usersCacheKeys  []string
	departmentsKeys []string
}

func (st *LoadUserSuite) SetupTest() {
	st.users = []*User{
		RandomUser(),
		RandomUser(),
		RandomUser(),
	}

	var sameDepUsers []*User
	st.departmentsKeys = nil
	for _, u := range st.users {
		anotherUser := RandomUser()
		anotherUser.Department = u.Department
		sameDepUsers = append(sameDepUsers, anotherUser)
		st.departmentsKeys = append(st.departmentsKeys, userByDepartmentCacheKey(u.Department))
	}

	st.users = append(st.users, sameDepUsers...)

	st.usersCacheKeys = nil
	for _, u := range st.users {
		st.usersCacheKeys = append(st.usersCacheKeys, userByIDCacheKey(u.ID))
	}
	st.storeUsers(st.users...)
}

func (st *LoadUserSuite) storeUsers(users ...*User) {
	st.T().Helper()
	var itemsToCache []*cache.Item
	for _, u := range users {
		itemsToCache = append(
			itemsToCache,
			&cache.Item{
				Key:   userByIDCacheKey(u.ID),
				Value: u,
			},
			&cache.Item{
				Key:   userByDepartmentCacheKey(u.Department),
				Field: string(u.ID),
				Value: u,
			},
		)
	}
	st.Require().NoError(
		st.cache.Set(st.ctx, itemsToCache...),
		"No error expected on preparing test data",
	)
}

func (st *LoadUserSuite) TestLoadSingleUser() {
	var dst *User

	err := st.cache.Get(st.ctx, &dst, userByIDCacheKey(st.users[0].ID))

	st.Require().NoError(err, "No error expected on loading from cache")
	st.Require().EqualValues(st.users[0], dst, "Unexpected user")
}

func (st *LoadUserSuite) TestLoadSeveralUsersIntoASlice() {
	var dst []*User

	err := st.cache.Get(st.ctx, &dst, st.usersCacheKeys...)

	st.Require().NoError(err, "No error expected on loading from cache")
	st.Require().Empty(cmp.Diff(st.users, dst), "non-matched users loaded")
}

func (st *LoadUserSuite) TestLoadSeveralUsersIntoAMap() {
	var dst map[string]*User

	err := st.cache.Get(st.ctx, &dst, st.usersCacheKeys...)

	st.Require().NoError(err, "No error expected on loading from cache")
	for idx, u := range st.users {
		k := st.usersCacheKeys[idx]
		st.Require().EqualValues(u, dst[k], "Unexpected user")
	}
}

func (st *LoadUserSuite) TestLoadSeveralUsersIntoAMap_WithKeyModification() {
	var dst map[string]*User

	err := st.cache.
		TransformCacheKeyForDestination(func(key, _ string, _ interface{}) (string, string, bool) {
			var userID string
			cachekeys.UnpackKey(key, &userID)
			return userID, "", false
		}).
		Get(st.ctx, &dst, st.usersCacheKeys...)

	st.Require().NoError(err, "No error expected on loading from cache")
	for _, u := range st.users {
		k := string(u.ID)
		st.Require().EqualValues(u, dst[k], "Unexpected user")
	}
}

func (st *LoadUserSuite) TestLoadAllItemsFromHashMap() {
	var dst []*User
	err := st.cache.HGetAll(st.ctx, &dst, st.departmentsKeys...)

	st.Require().NoError(err, "No error expected on loading from cache")
	st.Require().Empty(cmp.Diff(st.users, dst, cmpopts.SortSlices(func(a, b *User) bool {
		return a.ID > b.ID
	})), "non-matched users loaded")
}

func (st *LoadUserSuite) TestLoadSpecificFields() {
	var dst map[string]*User

	department := st.users[0].Department
	key := userByDepartmentCacheKey(department)
	var fields []string
	expectedMap := map[string]*User{}
	for _, u := range st.users {
		if department == u.Department {
			f := string(u.ID)
			fields = append(fields, f)
			expectedMap[cachekeys.KeyWithField(key, f)] = u
		}
	}

	err := st.cache.HGetFieldsForKey(st.ctx, &dst, key, fields...)

	st.Require().NoError(err, "No error expected on loading from cache")
	st.Require().Empty(cmp.Diff(expectedMap, dst, cmpopts.SortMaps(func(a, b *User) bool {
		return a.ID > b.ID
	})), "non-matched users loaded")
}

func TestLoadUserSuite(t *testing.T) {
	suite.Run(t, &LoadUserSuite{})
}

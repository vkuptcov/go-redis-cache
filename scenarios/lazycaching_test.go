package scenarios

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/suite"

	cache "github.com/vkuptcov/go-redis-cache/v8"
	"github.com/vkuptcov/go-redis-cache/v8/cachekeys"
)

type LazyCachingSuite struct {
	BaseCacheSuite
	users     []*User
	usersByID map[UserID]*User
}

func (st *LazyCachingSuite) SetupTest() {
	st.users = []*User{
		RandomUser(),
		RandomUser(),
		RandomUser(),
	}

	st.usersByID = map[UserID]*User{}

	for _, u := range st.users {
		st.usersByID[u.ID] = u
	}
}

func (st *LazyCachingSuite) TestLazyCacheSingleUser() {
	user := RandomUser()
	cacheKey := userByIDCacheKey(user.ID)

	var dst *User
	err := st.cache.
		WithAbsentKeysLoader(func(absentKeys ...string) (interface{}, error) {
			st.T().Helper()
			st.Require().EqualValues(cacheKey, absentKeys[0], "Unexpected cache key")
			return user, nil
		}).
		Get(
			st.ctx,
			&dst,
			cacheKey,
		)

	st.Require().NoError(err, "No error expected")
	st.Require().EqualValues(user, dst, "Unexpected user loaded")
	st.verifyUserPresenceInCache(user)
}

func (st *LazyCachingSuite) TestLazyCacheSeveralUsers_LoadKeyValuesMap() {
	var cacheKeys []string
	for _, u := range st.users {
		cacheKeys = append(cacheKeys, userByIDCacheKey(u.ID))
	}

	var dst []*User
	err := st.cache.
		WithAbsentKeysLoader(func(absentKeys ...string) (interface{}, error) {
			keyToUser := make(map[string]*User, len(absentKeys))
			for _, k := range absentKeys {
				var userID string
				cachekeys.UnpackKey(k, &userID)
				keyToUser[k] = st.usersByID[UserID(userID)]
			}
			return keyToUser, nil
		}).
		Get(
			st.ctx,
			&dst,
			cacheKeys...,
		)

	st.Require().NoError(err, "No error expected")
	st.Require().Empty(cmp.Diff(st.users, dst, cmpopts.SortSlices(func(a, b *User) bool {
		return a.ID > b.ID
	})), "non-matched users loaded")

	st.verifyUserPresenceInCache(st.users...)
}

func (st *LazyCachingSuite) TestLazyCacheSeveralUsers_LoadASliceOfUsers() {
	var cacheKeys []string
	for _, u := range st.users {
		cacheKeys = append(cacheKeys, userByIDCacheKey(u.ID))
	}

	var dst []*User
	err := st.cache.
		ExtractCacheKeyWith(func(it interface{}) (key, field string) {
			u := it.(*User)
			return userByIDCacheKey(u.ID), ""
		}).
		WithAbsentKeysLoader(func(absentKeys ...string) (interface{}, error) {
			users := make([]*User, 0, len(absentKeys))
			for _, k := range absentKeys {
				var userID string
				cachekeys.UnpackKey(k, &userID)
				if u, exists := st.usersByID[UserID(userID)]; exists {
					users = append(users, u)
				}
			}
			return users, nil
		}).
		Get(
			st.ctx,
			&dst,
			cacheKeys...,
		)

	st.Require().NoError(err, "No error expected")
	st.Require().Empty(cmp.Diff(st.users, dst, cmpopts.SortSlices(func(a, b *User) bool {
		return a.ID > b.ID
	})), "non-matched users loaded")

	st.verifyUserPresenceInCache(st.users...)
}

func (st *LazyCachingSuite) TestLazyCacheUsersByDepartmentsAndByIDs() {
	var cacheKeys []string
	for _, u := range st.users {
		cacheKeys = append(cacheKeys, userByDepartmentCacheKey(u.Department))
	}

	var dst map[string]*User
	err := st.cache.
		WithAbsentKeysLoader(func(absentKeys ...string) (interface{}, error) {
			var items []*cache.Item
			for _, k := range absentKeys {
				var department string
				cachekeys.UnpackKey(k, &department)
				for _, u := range st.users {
					if u.Department == department {
						// IMPORTANT:
						// As we need to cache a user by 2 ID and by department, we added 2 items here
						items = append(
							items,
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
				}
			}
			return items, nil
		}).
		TransformCacheKeyForDestination(func(key, field string, val interface{}) (newKey, newField string, skip bool) {
			if field != "" { // it means department part is absent in the key
				return "", "", true
			}
			var userID string
			cachekeys.UnpackKey(key, &userID)
			return userID, "", false
		}).
		Get(
			st.ctx,
			&dst,
			cacheKeys...,
		)

	st.Require().NoError(err, "No error expected")

	expectedMap := map[string]*User{}
	for uID, u := range st.usersByID {
		expectedMap[string(uID)] = u
	}

	st.Require().Empty(cmp.Diff(expectedMap, dst, cmpopts.SortMaps(func(a, b *User) bool {
		return a.ID > b.ID
	})), "non-matched users loaded")

	st.verifyUserPresenceInCache(st.users...)
	st.verifyUserByDepartmentPresenceInCache(st.users...)
}

func TestLazyCachingSuite(t *testing.T) {
	suite.Run(t, &LazyCachingSuite{})
}

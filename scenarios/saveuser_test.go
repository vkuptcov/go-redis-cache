package scenarios

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"syreclabs.com/go/faker"

	cache "github.com/vkuptcov/go-redis-cache/v8"
)

type SaveUserSuite struct {
	BaseCacheSuite
}

func (st *SaveUserSuite) TestSaveUser_ByID() {
	st.Run("save one user", func() {
		user := RandomUser()

		err := st.cache.SetKV(st.ctx, userByIDCacheKey(user.ID), user)

		st.Require().NoError(err, "save user failed")
		st.verifyUserPresenceInCache(user)
	})

	st.Run("save several users", func() {
		users := []*User{RandomUser(), RandomUser()}

		err := st.cache.SetKV(st.ctx,
			userByIDCacheKey(users[0].ID), users[0],
			userByIDCacheKey(users[1].ID), users[1],
		)

		st.Require().NoError(err, "save user failed")
		st.verifyUserPresenceInCache(users...)
	})

	st.Run("save several with explicit items creation", func() {
		users := []*User{RandomUser(), RandomUser()}
		err := st.cache.Set(
			st.ctx,
			&cache.Item{
				Key:   userByIDCacheKey(users[0].ID),
				Value: users[0],
			},
			&cache.Item{
				Key:   userByIDCacheKey(users[1].ID),
				Value: users[1],
			},
		)
		st.Require().NoError(err, "save user failed")
		st.verifyUserPresenceInCache(users...)
	})
}

func (st *SaveUserSuite) TestSaveUser_InHashMapByDepartment() {
	st.Run("save users to same hash map", func() {
		users := []*User{RandomUser(), RandomUser()}
		department := users[0].Department
		users[1].Department = department

		err := st.cache.HSetKV(
			st.ctx,
			userByDepartmentCacheKey(department), // common hash map key
			// hash map fields-values pairs
			string(users[0].ID), users[0],
			string(users[1].ID), users[1],
		)

		st.Require().NoError(err, "save user failed")
		st.verifyUserByDepartmentPresenceInCache(users...)
	})

	st.Run("save users to different hash maps", func() {
		users := []*User{RandomUser(), RandomUser()}

		err := st.cache.Set(
			st.ctx,
			&cache.Item{
				Key:   userByDepartmentCacheKey(users[0].Department),
				Field: string(users[0].ID),
				Value: users[0],
			},
			&cache.Item{
				Key:   userByDepartmentCacheKey(users[1].Department),
				Field: string(users[1].ID),
				Value: users[1],
			},
		)

		st.Require().NoError(err, "save user failed")
		st.verifyUserByDepartmentPresenceInCache(users...)
	})
}

func (st *SaveUserSuite) TestSaveUser_InBothHashMapAndByID() {
	user := RandomUser()

	err := st.cache.Set(
		st.ctx,
		&cache.Item{
			Key:   userByIDCacheKey(user.ID),
			Value: user,
		},
		&cache.Item{
			Key:   userByDepartmentCacheKey(user.Department),
			Field: string(user.ID),
			Value: user,
		},
	)

	st.Require().NoError(err, "save user failed")
	st.verifyUserPresenceInCache(user)
	st.verifyUserByDepartmentPresenceInCache(user)
}

func (st *SaveUserSuite) TestSaveUser_ConditionalCases() {
	st.Run("Save user if it only exists", func() {
		user := RandomUser()

		err := st.cache.Set(
			st.ctx,
			&cache.Item{
				Key:      userByIDCacheKey(user.ID),
				Value:    user,
				IfExists: true,
			},
		)
		st.Require().NoError(err, "save user failed")

		st.verifyUserAbsenceInCache(user)
	})

	st.Run("Override user in cache", func() {
		user := RandomUser()

		err := st.cache.Set(
			st.ctx,
			&cache.Item{
				Key:   userByIDCacheKey(user.ID),
				Value: user,
			},
		)
		st.Require().NoError(err, "save user failed")
		user.Name = faker.Name().FirstName()

		err = st.cache.Set(
			st.ctx,
			&cache.Item{
				Key:      userByIDCacheKey(user.ID),
				Value:    user,
				IfExists: true,
			},
		)
		st.Require().NoError(err, "save user failed")
		st.verifyUserPresenceInCache(user)
	})

	st.Run("Do not override user in cache", func() {
		initialUser := RandomUser()

		err := st.cache.Set(
			st.ctx,
			&cache.Item{
				Key:   userByIDCacheKey(initialUser.ID),
				Value: initialUser,
			},
		)
		st.Require().NoError(err, "save initialUser failed")

		updatedUser := *initialUser
		updatedUser.Name = faker.Name().FirstName()

		err = st.cache.Set(
			st.ctx,
			&cache.Item{
				Key:         userByIDCacheKey(initialUser.ID),
				Value:       &updatedUser,
				IfNotExists: true,
			},
		)
		st.Require().NoError(err, "user update failed")
		st.verifyUserPresenceInCache(initialUser)
	})

	st.Run("Do not override user in hash map", func() {
		initialUser := RandomUser()

		err := st.cache.Set(
			st.ctx,
			&cache.Item{
				Key:   userByDepartmentCacheKey(initialUser.Department),
				Field: string(initialUser.ID),
				Value: initialUser,
			},
		)
		st.Require().NoError(err, "save initialUser failed")

		updatedUser := *initialUser
		updatedUser.Name = faker.Name().FirstName()

		err = st.cache.Set(
			st.ctx,
			&cache.Item{
				Key:         userByDepartmentCacheKey(initialUser.Department),
				Field:       string(initialUser.ID),
				Value:       &updatedUser,
				IfNotExists: true,
			},
		)
		st.Require().NoError(err, "user update failed")
		st.verifyUserByDepartmentPresenceInCache(initialUser)
	})
}

func (st *SaveUserSuite) TestSaveUser_WithCustomTTL() {
	getTTL := func(k string) time.Duration {
		st.T().Helper()
		durationCmd := st.client.TTL(st.ctx, k)
		st.Require().NoError(durationCmd.Err(), "No error expected on getting TTL")
		return durationCmd.Val()
	}

	testCases := []struct {
		testCase    string
		setFunc     func(user *User) error
		expectedTTL time.Duration
	}{
		{
			testCase: "default TTL is taken from cache.DefaultTTL",
			setFunc: func(user *User) error {
				return st.cache.SetKV(st.ctx, userByIDCacheKey(user.ID), user)
			},
			expectedTTL: cache.DefaultDuration,
		},
		{
			testCase: "override TTL with WithTTL function",
			setFunc: func(user *User) error {
				return st.cache.
					WithTTL(1*time.Minute).
					SetKV(st.ctx, userByIDCacheKey(user.ID), user)
			},
			expectedTTL: 1 * time.Minute,
		},
		{
			testCase: "override TTL with Item",
			setFunc: func(user *User) error {
				return st.cache.
					WithTTL(1*time.Minute).
					Set(st.ctx, &cache.Item{
						Key:   userByIDCacheKey(user.ID),
						Value: user,
						TTL:   5 * time.Second,
					})
			},
			expectedTTL: 5 * time.Second,
		},
		{
			testCase: "store items without TTL",
			setFunc: func(user *User) error {
				return st.cache.
					WithTTL(-5*time.Second).
					SetKV(st.ctx, userByIDCacheKey(user.ID), user)
			},
			expectedTTL: -1 * time.Nanosecond,
		},
	}

	for _, tc := range testCases {
		st.Run(tc.testCase, func() {
			user := RandomUser()
			err := tc.setFunc(user)
			st.Require().NoError(err, "user save failed")

			ttl := getTTL(userByIDCacheKey(user.ID))

			st.Require().InDelta(tc.expectedTTL, ttl, float64(500*time.Millisecond), "Unexpected ttl")
		})
	}
}

func TestBasicCasesSuite(t *testing.T) {
	suite.Run(t, &SaveUserSuite{})
}

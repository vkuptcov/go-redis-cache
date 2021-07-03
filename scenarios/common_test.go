package scenarios

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"syreclabs.com/go/faker"

	cache "github.com/vkuptcov/go-redis-cache/v8"
	"github.com/vkuptcov/go-redis-cache/v8/cachekeys"
	"github.com/vkuptcov/go-redis-cache/v8/marshallers"
)

type UserID string

type User struct {
	ID         UserID
	Name       string
	Department string
}

func RandomUser() *User {
	return &User{
		ID:         UserID(faker.RandomString(5)),
		Name:       faker.Name().FirstName(),
		Department: faker.Commerce().Department() + "_" + faker.RandomString(3),
	}
}

const (
	cacheUserIDPrefix     = "usr-by-id"
	cacheDepartmentPrefix = "usr-by-dpmt"
)

func userByIDCacheKey(userID UserID) string {
	return cachekeys.CreateKey(cacheUserIDPrefix, string(userID))
}

func userByDepartmentCacheKey(department string) string {
	return cachekeys.CreateKey(cacheDepartmentPrefix, department)
}

type BaseCacheSuite struct {
	suite.Suite
	client     *redis.Client
	cache      *cache.Cache
	marshaller cache.Marshaller
	ctx        context.Context
}

func (st *BaseCacheSuite) SetupSuite() {
	st.client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	st.ctx = context.Background()

	st.marshaller = marshallers.NewMarshaller(&marshallers.JSONMarshaller{})

	st.cache = cache.NewCache(cache.Options{
		Redis:      st.client,
		DefaultTTL: 0,
		Marshaller: st.marshaller,
	})
}

func (st *BaseCacheSuite) verifyUserPresenceInCache(users ...*User) {
	st.T().Helper()
	var dst []*User
	var keys []string
	for _, u := range users {
		keys = append(keys, userByIDCacheKey(u.ID))
	}
	st.Require().NoError(
		st.cache.Get(st.ctx, &dst, keys...),
		"get user verification err",
	)
	st.Require().ElementsMatch(users, dst, "Unmatched users")
}

func (st *BaseCacheSuite) verifyUserAbsenceInCache(user *User) {
	st.T().Helper()
	var dst *User

	err := st.cache.
		AddCacheMissErrors().
		Get(st.ctx, &dst, userByIDCacheKey(user.ID))

	var keyErr *cache.KeyErr
	st.Require().True(errors.As(err, &keyErr), "KeyErr expected")

	st.Require().Nil(dst, "Dst must be nil")
}

func (st *BaseCacheSuite) verifyUserByDepartmentPresenceInCache(users ...*User) {
	st.T().Helper()
	for _, u := range users {
		var dst *User
		st.Require().NoError(
			st.cache.HGetFieldsForKey(st.ctx, &dst, userByDepartmentCacheKey(u.Department), string(u.ID)),
			"get user hash map verification err",
		)
		st.Require().EqualValues(u, dst, "Unmatched users")
	}
}

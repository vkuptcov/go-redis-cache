package examples

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/suite"
	"syreclabs.com/go/faker"

	cache "github.com/vkuptcov/go-redis-cache/v8"
	"github.com/vkuptcov/go-redis-cache/v8/cachekeys"
	"github.com/vkuptcov/go-redis-cache/v8/internal/marshaller"
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
		Department: faker.Commerce().Department(),
	}
}

func userByIDCacheKey(userID UserID) string {
	const cachePrefix = "usr-by-id"
	return cachekeys.CreateKey(cachePrefix, string(userID))
}

func userByDepartmentCacheKey(department string) string {
	const cachePrefix = "usr-by-dpmt"
	return cachekeys.CreateKey(cachePrefix, department)
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

	st.marshaller = marshaller.NewMarshaller(&marshallers.JSONMarshaller{})

	st.cache = cache.NewCache(cache.Options{
		Redis:      st.client,
		DefaultTTL: 0,
		Marshaller: st.marshaller,
	})
}

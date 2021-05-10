package examples

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	cache "github.com/vkuptcov/go-redis-cache/v8"
	"github.com/vkuptcov/go-redis-cache/v8/cachekeys"
	"github.com/vkuptcov/go-redis-cache/v8/internal/marshaller"
	"github.com/vkuptcov/go-redis-cache/v8/marshallers"
)

type User struct {
	ID         string
	Name       string
	Department string
}

func Example_saveAndGetUser() {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	cacheInst := cache.NewCache(cache.Options{
		Redis:      client,
		Marshaller: marshaller.NewMarshaller(&marshallers.JSONMarshaller{}),
	})

	user := &User{
		ID:         "u-1",
		Name:       "FirstUserName",
		Department: "R&D",
	}

	keyByID := cachekeys.CreateKey("usr", user.ID)
	keyByDepartment := cachekeys.CreateKey("usr-by-department", user.Department)

	saveErr := cacheInst.Set(
		context.Background(),
		&cache.Item{
			Key:   keyByID,
			Value: user,
		},
		&cache.Item{
			Key:   keyByDepartment,
			Field: user.ID,
			Value: user,
		},
	)
	if saveErr != nil {
		panic(saveErr)
	}

	var loadedUserByID User
	var loadedUserByDepartment User
	loadByIDErr := cacheInst.Get(
		context.Background(),
		&loadedUserByID,
		keyByID,
	)
	if loadByIDErr != nil {
		panic(loadByIDErr)
	}

	loadByDepartmentErr := cacheInst.HGetFieldsForKey(
		context.Background(),
		&loadedUserByDepartment,
		keyByDepartment,
		user.ID,
	)
	if loadByDepartmentErr != nil {
		panic(loadByDepartmentErr)
	}

	fmt.Println(loadedUserByID)
	fmt.Println(loadedUserByDepartment)

	// Output: {u-1 FirstUserName R&D}
	// {u-1 FirstUserName R&D}
}

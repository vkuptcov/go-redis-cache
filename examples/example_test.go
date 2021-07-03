package examples

import (
	"context"
	"fmt"
	"sort"

	"github.com/go-redis/redis/v8"
	cache "github.com/vkuptcov/go-redis-cache/v8"
	"github.com/vkuptcov/go-redis-cache/v8/cachekeys"
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
		Marshaller: marshallers.NewMarshaller(&marshallers.JSONMarshaller{}),
	})

	user := &User{
		ID:         "u-1",
		Name:       "FirstUserName",
		Department: "R&D",
	}

	keyByID := cachekeys.CreateKey("usr", user.ID)
	keyByDepartment := cachekeys.CreateKey("usr-by-department", user.Department)

	// store user with a key derived for their id
	// store the same user with a key derived from department
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
	// get user by id
	loadByIDErr := cacheInst.Get(
		context.Background(),
		&loadedUserByID,
		keyByID,
	)
	if loadByIDErr != nil {
		panic(loadByIDErr)
	}

	// get user by department
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
	fmt.Println(loadedUserByDepartment == loadedUserByID)

	// Output: {u-1 FirstUserName R&D}
	// {u-1 FirstUserName R&D}
	// true
}

func Example_saveSeveralUsersAndLoadInSlice() {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	cacheInst := cache.NewCache(cache.Options{
		Redis:      client,
		Marshaller: marshallers.NewMarshaller(&marshallers.JSONMarshaller{}),
	})

	keyByID := func(id string) string {
		return cachekeys.CreateKey("usr", id)
	}

	firstUser := &User{
		ID:         "u-1",
		Name:       "FirstUserName",
		Department: "R&D",
	}
	secondUser := &User{
		ID:         "u-2",
		Name:       "SecondUserName",
		Department: "IT",
	}

	saveErr := cacheInst.Set(
		context.Background(),
		&cache.Item{
			Key:   keyByID(firstUser.ID),
			Value: firstUser,
		},
		&cache.Item{
			Key:   keyByID(secondUser.ID),
			Value: secondUser,
		},
	)
	if saveErr != nil {
		panic(saveErr)
	}

	var loadedUsers []*User

	loadErr := cacheInst.Get(
		context.Background(),
		&loadedUsers,
		keyByID(firstUser.ID),
		keyByID(secondUser.ID),
	)
	if loadErr != nil {
		panic(loadErr)
	}
	sort.Slice(loadedUsers, func(i, j int) bool {
		return loadedUsers[i].ID < loadedUsers[j].ID
	})
	for _, u := range loadedUsers {
		fmt.Println(u)
	}
	// Output:&{u-1 FirstUserName R&D}
	// &{u-2 SecondUserName IT}
}

func Example_loadUsers() {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	cacheInst := cache.NewCache(cache.Options{
		Redis:      client,
		Marshaller: marshallers.NewMarshaller(&marshallers.JSONMarshaller{}),
	})

	keyByID := func(id string) string {
		return cachekeys.CreateKey("usr", id)
	}

	var loadedUsers []*User

	loadErr := cacheInst.
		WithAbsentKeysLoader(func(absentKeys ...string) (interface{}, error) {
			users := make([]*User, 0, len(absentKeys))
			for _, k := range absentKeys {
				var userID string
				cachekeys.UnpackKey(k, &userID)
				users = append(users, &User{
					ID:         userID,
					Name:       userID + "-name",
					Department: userID + "-dep",
				})
			}
			return users, nil
		}).
		ExtractCacheKeyWith(func(it interface{}) (key, field string) {
			// we need this function as the loader function return a slice of objects
			// so we need to create cache keys out of them
			// this function isn't needed in case we return a slice of *cache.Item or
			// a map of keys to a loaded item
			return keyByID(it.(*User).ID), ""
		}).
		Get(
			context.Background(),
			&loadedUsers,
			keyByID("1"),
			keyByID("2"),
		)

	sort.Slice(loadedUsers, func(i, j int) bool {
		return loadedUsers[i].ID < loadedUsers[j].ID
	})
	for _, u := range loadedUsers {
		fmt.Println(u)
	}

	if loadErr != nil {
		panic(loadErr)
	}
	// Output:&{1 1-name 1-dep}
	// &{2 2-name 2-dep}
}

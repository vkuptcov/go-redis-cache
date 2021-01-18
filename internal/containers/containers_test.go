package containers

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
	"syreclabs.com/go/faker"
)

type ContainersSuite struct {
	expectedSliceContainer sliceContainer
	expectedMapContainer   mapContainer
	suite.Suite
}

func (st *ContainersSuite) TestContainerCreation() {
	st.Run("create a container for a nil slice", func() {
		var dst []string
		c, err := NewContainer(&dst)
		st.Require().NoError(err, "No error expected on container creation")
		st.Require().IsType(st.expectedSliceContainer, c)
	})
	st.Run("create a container for an empty slice", func() {
		dst := []string{}
		c, err := NewContainer(&dst)
		st.Require().NoError(err, "No error expected on container creation")
		st.Require().IsType(st.expectedSliceContainer, c)
	})

	st.Run("create a container for a nil map", func() {
		var dst map[string]string
		c, err := NewContainer(&dst)
		st.Require().NoError(err, "No error expected on container creation")
		st.Require().IsType(st.expectedMapContainer, c)
	})
	st.Run("create a container for an empty map", func() {
		dst := map[string]string{}
		c, err := NewContainer(&dst)
		st.Require().NoError(err, "No error expected on container creation")
		st.Require().IsType(st.expectedMapContainer, c)
	})
}

func (st *ContainersSuite) TestAddElementsIntoSliceContainer() {
	var dst []string
	var expectedSlice []string
	c, err := NewContainer(&dst)
	st.Require().NoError(err, "No error expected on container creation")
	for i := 0; i < 5; i++ {
		k := faker.RandomString(5)
		v := faker.Lorem().Sentence(2)
		c.AddElement(k, &v)
		expectedSlice = append(expectedSlice, v)
		st.Run(strconv.Itoa(i), func() {
			st.Require().EqualValues(expectedSlice, dst)
		})
	}
}

func (st *ContainersSuite) TestAddElementsIntoMapContainer() {
	var dst map[string]string
	expectedMap := map[string]string{}
	c, err := NewContainer(&dst)
	st.Require().NoError(err, "No error expected on container creation")
	c.InitWithSize(0)
	for i := 0; i < 5; i++ {
		k := faker.RandomString(5)
		v := faker.Lorem().Sentence(2)
		c.AddElement(k, &v)
		expectedMap[k] = v
		st.Run(strconv.Itoa(i), func() {
			st.Require().EqualValues(expectedMap, dst)
		})
	}
}

func TestContainersSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &ContainersSuite{})
}

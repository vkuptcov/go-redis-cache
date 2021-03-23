package marshaller

import (
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/vkuptcov/go-redis-cache/v8/marshallers"
)

type structureToSerialize struct {
	Field string
}

type MarshalUnMarshalSuite struct {
	marshaller *baseMarshaller
	suite.Suite
}

func (st *MarshalUnMarshalSuite) SetupSuite() {
	st.marshaller = &baseMarshaller{
		customMarshaller: &marshallers.JSONMarshaller{},
	}
}

var marshallerTestData = []struct {
	testCase     string
	unmarshalled interface{}
	marshalled   string
}{
	{
		testCase:     "String",
		unmarshalled: "some string",
		marshalled:   "some string",
	},
	{
		testCase:     "Bytes",
		unmarshalled: []byte{49, 50, 51, 52, 53},
		marshalled:   "12345",
	},
	{
		testCase:     "Int",
		unmarshalled: -11,
		marshalled:   "-11",
	},
	{
		testCase:     "Int8",
		unmarshalled: int8(-1),
		marshalled:   "-1",
	},
	{
		testCase:     "Int16",
		unmarshalled: int16(math.MinInt8) - 1,
		marshalled:   "-129",
	},
	{
		testCase:     "Int32",
		unmarshalled: int32(math.MinInt16) - 1,
		marshalled:   "-32769",
	},
	{
		testCase:     "Int64",
		unmarshalled: int64(math.MaxInt32) + 1,
		marshalled:   "2147483648",
	},
	{
		testCase:     "Float32",
		unmarshalled: float32(0.17),
		marshalled:   "0.17",
	},
	{
		testCase:     "Float64",
		unmarshalled: 0.00021,
		marshalled:   "0.00021",
	},
	{
		testCase:     "Bool false",
		unmarshalled: false,
		marshalled:   "f",
	},
	{
		testCase:     "Bool true",
		unmarshalled: true,
		marshalled:   "t",
	},
	// @todo fix interface unmarshal part
	//nolint:gocritic // it complains about no-spaces between code and comment
	//{
	//	testCase:     "Struct",
	//	unmarshalled: &structureToSerialize{Field: "f1"},
	//	marshalled:   "{\"Field\":\"f1\"}",
	//},
}

func (st *MarshalUnMarshalSuite) Test_Marshal() {
	for _, td := range marshallerTestData {
		st.Run(td.testCase, func() {
			result, marshalErr := st.marshaller.Marshal(td.unmarshalled)
			st.Require().NoError(marshalErr, "No marshal error expected")
			st.Require().EqualValues(td.marshalled, string(result), "Unexpected marshal result")
		})
	}
}

func (st *MarshalUnMarshalSuite) Test_Unmarshal() {
	for _, td := range marshallerTestData {
		st.Run(td.testCase, func() {
			dst := reflect.New(reflect.Indirect(reflect.ValueOf(td.unmarshalled)).Type()).Elem().Interface()
			unmarshalErr := st.marshaller.Unmarshal([]byte(td.marshalled), &dst)
			st.Require().NoError(unmarshalErr, "No marshal error expected")
			st.Require().EqualValues(td.unmarshalled, dst, "values must match")
		})
	}
}

func (st *MarshalUnMarshalSuite) Test_Marshal_Nil() {
	result, marshalErr := st.marshaller.Marshal(nil)
	st.Require().NoError(marshalErr, "No marshal error expected")
	st.Require().Nil(result, "unexpected result")
}

func (st *MarshalUnMarshalSuite) Test_Unmarshal_Nil() {
	unmarshalErr := st.marshaller.Unmarshal([]byte(""), nil)
	st.Require().NoError(unmarshalErr, "No marshal error expected")
}

func (st *MarshalUnMarshalSuite) Test_Unmarshal_String() {
	var dst string
	expected := "test string"
	unmarshalErr := st.marshaller.Unmarshal([]byte(expected), &dst)
	st.Require().NoError(unmarshalErr, "No unmarshal error expected")
	st.Require().Equal(expected, dst, "Unexpected unmarshalled result")
}

func (st *MarshalUnMarshalSuite) Test_Unmarshal_Struct() {
	var dst *structureToSerialize
	expected := &structureToSerialize{Field: "f1"}
	unmarshalErr := st.marshaller.Unmarshal([]byte("{\"Field\":\"f1\"}"), &dst)
	st.Require().NoError(unmarshalErr, "No unmarshal error expected")
	st.Require().Equal(expected, dst, "Unexpected unmarshalled result")
}

func TestMarshalUnMarshalSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &MarshalUnMarshalSuite{})
}

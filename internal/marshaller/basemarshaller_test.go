package marshaller

import (
	"testing"

	"github.com/stretchr/testify/suite"
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
		customMarshaller: &JSONMarshaller{},
	}
}

func (st *MarshalUnMarshalSuite) Test_Marshal() {
	testData := []struct {
		testCase string
		val      interface{}
		expected interface{}
	}{
		{
			testCase: "Nil",
			val:      nil,
			expected: nil,
		},
		{
			testCase: "String",
			val:      "some string",
			expected: "some string",
		},
		{
			testCase: "Bytes",
			val:      []byte{49, 50, 51, 52, 53},
			expected: "12345",
		},
		{
			testCase: "Int",
			val:      11,
			expected: "11",
		},
		{
			testCase: "Int64",
			val:      int64(21),
			expected: "21",
		},
		{
			testCase: "Float",
			val:      0.17,
			expected: "0.17",
		},
		{
			testCase: "Struct",
			val:      &structureToSerialize{Field: "f1"},
			expected: "{\"Field\":\"f1\"}",
		},
	}
	for _, td := range testData {
		st.Run(td.testCase, func() {
			result, marshalErr := st.marshaller.Marshal(td.val)
			st.Require().NoError(marshalErr, "No marshal error expected")
			if td.expected != nil {
				st.Require().EqualValues(td.expected, string(result), "Unexpected marshal result")
			} else {
				st.Require().Nil(result, "Nil result expected")
			}
		})
	}
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

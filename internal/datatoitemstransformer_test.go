package internal

import (
	"testing"

	requireLib "github.com/stretchr/testify/require"
)

func TestNewDataTransformer(t *testing.T) {
	testCases := []struct {
		testCase            string
		data                interface{}
		expectedTransformer interface{}
	}{
		{
			testCase: "create a map transformer",
			data: map[string]string{
				"k": "v",
			},
			expectedTransformer: mapTransformer{},
		},
		{
			testCase:            "create a slice transformer",
			data:                []int{},
			expectedTransformer: sliceTransformer{},
		},
		{
			testCase:            "create a single transformer",
			data:                1,
			expectedTransformer: singleElementTransformer{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testCase, func(t *testing.T) {
			dt := newDataTransformer([]string{}, tc.data, nil)
			requireLib.New(t).IsType(tc.expectedTransformer, dt, "unexpected transformer type")
		})
	}
}

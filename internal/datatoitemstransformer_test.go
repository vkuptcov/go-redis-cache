package internal

import (
	"strings"
	"testing"

	requireLib "github.com/stretchr/testify/require"

	"github.com/vkuptcov/go-redis-cache/v8/cachekeys"
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

func TestDataTransformer_GetItems(t *testing.T) {
	testCases := []struct {
		testCase         string
		absentKeys       []string
		data             interface{}
		itemToCacheKeyFn func(it interface{}) (key, field string)

		expectedItems []*Item
	}{
		{
			testCase:   "transform single element",
			absentKeys: []string{"key"},
			data:       "value",
			expectedItems: []*Item{
				{
					Key:   "key",
					Value: "value",
				},
			},
		},
		{
			testCase:   "transform single element to a hash map item",
			absentKeys: []string{cachekeys.KeyWithField("key", "field")},
			data:       "value",
			expectedItems: []*Item{
				{
					Key:   "key",
					Field: "field",
					Value: "value",
				},
			},
		},
		{
			testCase:   "return single Item as is, ignoring the provided absent keys",
			absentKeys: []string{"whatever keys"},
			data: &Item{
				Key:   "key",
				Value: "value",
				Field: "field",
			},
			expectedItems: []*Item{
				{
					Key:   "key",
					Field: "field",
					Value: "value",
				},
			},
		},
		{
			testCase:   "transform single element with cache key transformation function",
			absentKeys: []string{"key"},
			data:       "value",
			itemToCacheKeyFn: func(it interface{}) (key, field string) {
				t.Helper()
				requireLib.New(t).EqualValues("value", it, "unexpected item provided")
				return "transformed-key", "transformed-value"
			},
			expectedItems: []*Item{
				{
					Key:   "transformed-key",
					Field: "transformed-value",
					Value: "value",
				},
			},
		},
		{
			testCase:   "return a list of elements",
			absentKeys: []string{"key1", "key2"},
			data:       []string{"val1", "val2"},
			itemToCacheKeyFn: func(it interface{}) (key, field string) {
				return strings.ReplaceAll(it.(string), "val", "key"), ""
			},
			expectedItems: []*Item{
				{
					Key:   "key1",
					Value: "val1",
				},
				{
					Key:   "key2",
					Value: "val2",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testCase, func(t *testing.T) {
			transformer := newDataTransformer(tc.absentKeys, tc.data, tc.itemToCacheKeyFn)

			items, getErr := transformer.getItems()

			require := requireLib.New(t)
			require.NoError(getErr, "No errors expected on getting items")
			require.ElementsMatch(tc.expectedItems, items)
		})
	}
}

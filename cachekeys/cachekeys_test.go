package cachekeys

import (
	"strings"
	"testing"

	requireLib "github.com/stretchr/testify/require"
)

var upackKeyData = []struct {
	key      string
	expected []string
	testCase string
}{
	{
		key:      "prefix|a|b|c|d",
		expected: []string{"prefix", "a", "b", "c", "d"},
		testCase: "all params are present",
	},
	{
		key:      "prefix|a|b|c/field",
		expected: []string{"prefix", "a", "b", "c", "field"},
		testCase: "all params with a field are present",
	},
	{
		key:      "prefix|a|b",
		expected: []string{"prefix", "a", "b", "", ""},
		testCase: "more params then the keys are provided",
	},
	{
		key:      "prefix|a|b|c|d",
		expected: []string{"prefix", "a", "b"},
		testCase: "less params then the keys are provided",
	},
	// @todo implement it
	// {
	//	key:      "prefix|a|b/shouldNotBeInterpretedAsAField|c/field",
	//	expected: []string{"prefix", "a", "b/shouldNotBeInterpretedAsAField", "c", "field"},
	//	testCase: "all params with a field are present",
	// },
	{
		key:      "prefix",
		expected: []string{"prefix"},
		testCase: "only prefix is set",
	},
	{
		key:      "",
		expected: []string{},
		testCase: "empty string",
	},
}

func TestUnpackKeyWithPrefix(t *testing.T) {
	for _, td := range upackKeyData {
		t.Run(td.testCase, func(t *testing.T) {
			require := requireLib.New(t)
			key := td.key
			strs, pointers := makeStringsAndPointers(len(td.expected))
			UnpackKeyWithPrefix(key, pointers...)
			require.Equal(td.expected, strs)
		})
	}
}

func TestCreateAndUnpackKeyWithPrefix(t *testing.T) {
	testData := [][]string{
		{"prefix", "key1"},
		{"prefix", "key1", "key2", "key3"},
	}

	for _, td := range testData {
		t.Run(strings.Join(td, ", "), func(t *testing.T) {
			require := requireLib.New(t)
			key := CreateKey(td[0], td[1], td[2:]...)
			strsWithPrefix, pointersWithPrefixes := makeStringsAndPointers(len(td))
			UnpackKeyWithPrefix(key, pointersWithPrefixes...)
			require.EqualValues(td, strsWithPrefix, "Unmatched elements")

			strsWithoutPrefix, pointersWithoutPrefixes := makeStringsAndPointers(len(td) - 1)
			UnpackKey(key, pointersWithoutPrefixes...)
			require.EqualValues(td[1:], strsWithoutPrefix, "Unmatched elements")
		})
	}
}

func TestUnpackKey(t *testing.T) {
	for _, td := range upackKeyData {
		t.Run(td.testCase, func(t *testing.T) {
			require := requireLib.New(t)
			key := td.key
			expected := td.expected
			if len(td.expected) > 0 {
				expected = td.expected[1:]
			}
			strs, pointers := makeStringsAndPointers(len(expected))
			UnpackKey(key, pointers...)
			require.Equal(strs, expected)
		})
	}
}

func TestUnpackKeyWithPrefix_IgnoringPrefix(t *testing.T) {
	require := requireLib.New(t)
	var s string
	UnpackKeyWithPrefix("prefix|a|b", nil, &s, nil, nil)
	require.Equal(s, "a")
}

func makeStringsAndPointers(length int) (strs []string, pointers []*string) {
	strs = make([]string, length)
	pointers = make([]*string, length)
	for i := range strs {
		pointers[i] = &strs[i]
	}
	for i := range strs {
		pointers[i] = &strs[i]
	}
	return
}

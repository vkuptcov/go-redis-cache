package cachekeys

import (
	"strings"
)

const (
	keysSeparator  = "|"
	fieldSeparator = "/"
)

func CreateKey(prefix, firstKey string, compounds ...string) string {
	return prefix + keysSeparator + firstKey + keysSeparator + strings.Join(compounds, keysSeparator)
}

// UnpackKey extracts parts from a key, with prefix.
// E.g.
// var prefix, userID string
// UnpackKey("usr_by_id|123", &prefix, &userID)
// writes
// "usr_by_id" into the prefix var
// "123" into the userID var
func UnpackKeyWithPrefix(key string, parts ...*string) {
	key = strings.ReplaceAll(key, fieldSeparator, keysSeparator)
	for idx, s := range strings.Split(key, keysSeparator) {
		if idx >= len(parts) {
			break
		}
		if parts[idx] != nil {
			*parts[idx] = s
		}
	}
}

// UnpackKey extracts parts from a key, ignoring prefix.
// E.g.
// var userID string
// UnpackKey("usr_by_id|123", &userID)
// writes "123" into the userID
func UnpackKey(key string, parts ...*string) {
	prefixedSlice := append([]*string{nil}, parts...)
	UnpackKeyWithPrefix(key, prefixedSlice...)
}

func KeyWithField(key, field string) string {
	return key + fieldSeparator + field
}

func SplitKeyAndField(s string) (key, field string) {
	lastIndex := strings.LastIndex(s, fieldSeparator)
	key = s
	if lastIndex >= 0 {
		key = s[:lastIndex]
		field = s[lastIndex+1:]
	}
	return key, field
}

package cachekeys

import "strings"

const (
	keysSeparator  = "|"
	fieldSeparator = "/"
)

func CreateKey(prefix, firstKey string, compounds ...string) string {
	return prefix + keysSeparator + firstKey + keysSeparator + strings.Join(compounds, keysSeparator)
}

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

func UnpackKey(key string, parts ...*string) {
	prefixedSlice := append([]*string{nil}, parts...)
	UnpackKeyWithPrefix(key, prefixedSlice...)
}

func KeyWithField(key, field string) string {
	return key + fieldSeparator + field
}

func HasFieldInKey(key string) bool {
	return strings.Contains(key, fieldSeparator)
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

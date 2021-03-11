package keys

import "strings"

const (
	keysSeparator  = "|"
	fieldSeparator = "/"
)

func CreateKey(prefix, firstKey string, compounds ...string) string {
	return prefix + keysSeparator + firstKey + keysSeparator + strings.Join(compounds, keysSeparator)
}

func UnpackKeyWithPrefix(key string, parts ...*string) {
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

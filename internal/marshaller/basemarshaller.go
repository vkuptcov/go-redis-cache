package marshaller

import (
	"reflect"

	"github.com/vkuptcov/go-redis-cache/v8/internal"
)

type baseMarshaller struct {
	customMarshaller internal.Marshaller
}

func NewMarshaller(customMarshaller internal.Marshaller) internal.Marshaller {
	return &baseMarshaller{customMarshaller: customMarshaller}
}

// @todo implement int* default marshaller
func (m *baseMarshaller) Marshal(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	}

	return m.customMarshaller.Marshal(value)
}

func (m *baseMarshaller) Unmarshal(data []byte, dst interface{}) error {
	if len(data) == 0 {
		return nil
	}

	switch v := dst.(type) {
	case nil:
		return nil
	case *[]byte:
		clone := make([]byte, len(data))
		copy(clone, data)
		*v = clone
		return nil
	case *string:
		*v = string(data)
		return nil
	case *interface{}:
		t := reflect.Indirect(reflect.ValueOf(dst)).Elem().Kind()
		if t == reflect.String {
			*v = string(data)
			return nil
		}
	}

	return m.customMarshaller.Unmarshal(data, dst)
}

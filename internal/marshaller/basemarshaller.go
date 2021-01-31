package marshaller

import (
	"reflect"
	"strconv"

	"github.com/vkuptcov/go-redis-cache/v8/internal"
)

type baseMarshaller struct {
	customMarshaller internal.Marshaller
}

func NewMarshaller(customMarshaller internal.Marshaller) internal.Marshaller {
	return &baseMarshaller{customMarshaller: customMarshaller}
}

//nolint:gocyclo // we can't do a lot of here as it's better to use something faster than fmt.Sprintf
func (m *baseMarshaller) Marshal(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case int:
		return []byte(strconv.Itoa(v)), nil
	case int8:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int16:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int32:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int64:
		return []byte(strconv.FormatInt(v, 10)), nil
	case uint:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint8:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint16:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint32:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint64:
		return []byte(strconv.FormatUint(v, 10)), nil
	case float64:
		return []byte(strconv.FormatFloat(v, 'f', -1, 64)), nil
	case float32:
		return []byte(strconv.FormatFloat(float64(v), 'f', -1, 32)), nil
	case bool:
		if v {
			return []byte("t"), nil
		}
		return []byte("f"), nil
	}

	return m.customMarshaller.Marshal(value)
}

//nolint:gocyclo // here we need to enumerate all the options
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
	case *int:
		parsed, err := strconv.Atoi(string(data))
		*v = parsed
		return err
	case *int8:
		parsed, err := strconv.ParseInt(string(data), 10, 8)
		*v = int8(parsed)
		return err
	case *int16:
		parsed, err := strconv.ParseInt(string(data), 10, 16)
		*v = int16(parsed)
		return err
	case *int32:
		parsed, err := strconv.ParseInt(string(data), 10, 32)
		*v = int32(parsed)
		return err
	case *int64:
		parsed, err := strconv.ParseInt(string(data), 10, 64)
		*v = parsed
		return err
	case *uint:
		parsed, err := strconv.ParseUint(string(data), 10, 64)
		*v = uint(parsed)
		return err
	case *uint8:
		parsed, err := strconv.ParseUint(string(data), 10, 8)
		*v = uint8(parsed)
		return err
	case *uint16:
		parsed, err := strconv.ParseUint(string(data), 10, 16)
		*v = uint16(parsed)
		return err
	case *uint32:
		parsed, err := strconv.ParseUint(string(data), 10, 32)
		*v = uint32(parsed)
		return err
	case *uint64:
		parsed, err := strconv.ParseUint(string(data), 10, 64)
		*v = parsed
		return err
	case *float32:
		float, err := strconv.ParseFloat(string(data), 32)
		*v = float32(float)
		return err
	case *float64:
		float, err := strconv.ParseFloat(string(data), 64)
		*v = float
		return err
	case *bool:
		*v = string(data) == "t"
		return nil
	case *interface{}:
		// @todo try to unify it with the prev parts
		dd := reflect.Indirect(reflect.ValueOf(dst)).Interface()
		switch dd.(type) {
		case nil:
			return nil
		case []byte:
			clone := make([]byte, len(data))
			copy(clone, data)
			*v = clone
			return nil
		case string:
			*v = string(data)
			return nil
		case int:
			parsed, err := strconv.Atoi(string(data))
			*v = parsed
			return err
		case int8:
			parsed, err := strconv.ParseInt(string(data), 10, 8)
			*v = int8(parsed)
			return err
		case int16:
			parsed, err := strconv.ParseInt(string(data), 10, 16)
			*v = int16(parsed)
			return err
		case int32:
			parsed, err := strconv.ParseInt(string(data), 10, 32)
			*v = int32(parsed)
			return err
		case int64:
			parsed, err := strconv.ParseInt(string(data), 10, 64)
			*v = parsed
			return err
		case uint:
			parsed, err := strconv.ParseUint(string(data), 10, 64)
			*v = uint(parsed)
			return err
		case uint8:
			parsed, err := strconv.ParseUint(string(data), 10, 8)
			*v = uint8(parsed)
			return err
		case uint16:
			parsed, err := strconv.ParseUint(string(data), 10, 16)
			*v = uint16(parsed)
			return err
		case uint32:
			parsed, err := strconv.ParseUint(string(data), 10, 32)
			*v = uint32(parsed)
			return err
		case uint64:
			parsed, err := strconv.ParseUint(string(data), 10, 64)
			*v = parsed
			return err
		case float32:
			float, err := strconv.ParseFloat(string(data), 32)
			*v = float32(float)
			return err
		case float64:
			float, err := strconv.ParseFloat(string(data), 64)
			*v = float
			return err
		case bool:
			*v = string(data) == "t"
			return nil
		}
	}
	return m.customMarshaller.Unmarshal(data, dst)
}

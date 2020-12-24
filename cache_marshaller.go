package cache

type Marshaller interface {
	Marshal(value interface{}) ([]byte, error)
	Unmarshal(data []byte, dst interface{}) error
}

type baseMarshaller struct {
	customMarshaller Marshaller
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
	}

	return m.customMarshaller.Unmarshal(data, dst)
}

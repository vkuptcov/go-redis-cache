package marshallers

type Marshaller interface {
	Marshal(value interface{}) ([]byte, error)
	Unmarshal(data []byte, dst interface{}) error
}

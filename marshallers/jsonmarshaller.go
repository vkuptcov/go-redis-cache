package marshallers

import (
	"encoding/json"
)

type JSONMarshaller struct{}

func (t *JSONMarshaller) Marshal(val interface{}) ([]byte, error) {
	return json.Marshal(val)
}

func (t *JSONMarshaller) Unmarshal(data []byte, dst interface{}) error {
	return json.Unmarshal(data, dst)
}

var _ Marshaller = &JSONMarshaller{}

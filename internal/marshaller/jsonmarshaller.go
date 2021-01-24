package marshaller

import (
	"encoding/json"

	"github.com/vkuptcov/go-redis-cache/v8/internal"
)

type JSONMarshaller struct{}

func (t *JSONMarshaller) Marshal(val interface{}) ([]byte, error) {
	return json.Marshal(val)
}

func (t *JSONMarshaller) Unmarshal(data []byte, dst interface{}) error {
	return json.Unmarshal(data, dst)
}

var _ internal.Marshaller = &JSONMarshaller{}

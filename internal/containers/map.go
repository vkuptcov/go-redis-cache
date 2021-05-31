package containers

import (
	"reflect"

	"github.com/vkuptcov/go-redis-cache/v8/cachekeys"
)

type mapContainer struct {
	*baseContainer
}

func (m mapContainer) AddElementWithSubkey(key, subkey string, value interface{}) {
	if subkey != "" {
		key = cachekeys.KeyWithField(key, subkey)
	}
	m.AddElement(key, value)
}

func (m mapContainer) AddElement(key string, value interface{}) {
	m.cntValue.SetMapIndex(reflect.ValueOf(key), m.dstElementToValue(value))
}

func (m mapContainer) InitWithSize(size int) {
	if m.cntValue.IsNil() {
		m.cntValue.Set(reflect.MakeMapWithSize(m.cntType, size))
	}
}

var _ Container = mapContainer{}

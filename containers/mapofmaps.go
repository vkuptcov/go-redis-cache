package containers

import (
	"reflect"
)

type mapOfMapsContainer struct {
	*baseContainer
}

func (m mapOfMapsContainer) AddElement(key, field string, value interface{}) {
	keyValue := reflect.ValueOf(key)
	dstMap := m.cntValue.MapIndex(keyValue)
	if !dstMap.IsValid() || dstMap.IsNil() {
		dstMap = reflect.MakeMapWithSize(m.cntType.Elem(), 1)
	}
	dstMap.SetMapIndex(reflect.ValueOf(field), m.dstElementToValue(value))
	m.cntValue.SetMapIndex(keyValue, dstMap)
}

func (m mapOfMapsContainer) InitWithSize(size int) {
	if m.cntValue.IsNil() {
		m.cntValue.Set(reflect.MakeMapWithSize(m.cntType, size))
	}
}

var _ Container = mapOfMapsContainer{}

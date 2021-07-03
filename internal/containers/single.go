package containers

import (
	"reflect"
)

type singleElement struct {
	dst             interface{}
	assignableValue reflect.Value
}

func (s singleElement) DstEl() interface{} {
	return s.dst
}

func (s singleElement) AddElementWithSubkey(_, _ string, value interface{}) {
	s.AddElement("", value)
}

func (s singleElement) AddElement(_ string, value interface{}) {
	val := reflect.ValueOf(value)
	assignableType := s.assignableValue.Type()
	if assignableType.AssignableTo(val.Type()) {
		s.assignableValue.Set(reflect.ValueOf(value))
	} else {
		s.assignableValue.Set(reflect.Indirect(val))
	}
}

func (s singleElement) IsMultiElementContainer() bool {
	return false
}

func (s singleElement) InitWithSize(_ int) {}

var _ Container = singleElement{}

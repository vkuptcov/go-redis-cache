package containers

import (
	"reflect"
)

type SingleElement struct {
	dst             interface{}
	assignableValue reflect.Value
}

func (s SingleElement) DstEl() interface{} {
	return s.dst
}

func (s SingleElement) AddElementWithSubkey(_, _ string, value interface{}) {
	s.AddElement("", value)
}

func (s SingleElement) AddElement(_ string, value interface{}) {
	val := reflect.ValueOf(value)
	assignableType := s.assignableValue.Type()
	if assignableType.AssignableTo(val.Type()) {
		s.assignableValue.Set(reflect.ValueOf(value))
	} else {
		s.assignableValue.Set(reflect.Indirect(val))
	}
}

func (s SingleElement) InitWithSize(_ int) {}

var _ Container = SingleElement{}

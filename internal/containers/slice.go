package containers

import (
	"reflect"
)

type sliceContainer struct {
	*baseContainer
}

func (s sliceContainer) AddElement(_, _ string, value interface{}) {
	s.cntValue = reflect.Append(s.cntValue, s.dstElementToValue(value))
	s.assignableValue.Set(s.cntValue)
}

func (s sliceContainer) InitWithSize(size int) {
	if s.cntValue.IsNil() {
		s.cntValue.Set(reflect.MakeSlice(s.cntType, 0, size))
	}
}

var _ Container = sliceContainer{}

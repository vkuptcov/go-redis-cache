package containers

import (
	"reflect"

	"github.com/pkg/errors"
)

var (
	ErrNonContainerType = errors.New("dst must be a map or a slice")
)

type Container interface {
	DstEl() interface{}
	AddElementWithSubkey(key, subkey string, value interface{})
	AddElement(key string, value interface{})
	InitWithSize(size int)
}

type baseContainer struct {
	// assignableValue is different from cntValue in case of container element
	// is defined as an interface{}
	assignableValue   reflect.Value
	cntValue          reflect.Value
	cntType           reflect.Type
	elementType       reflect.Type
	isElementAPointer bool
}

func (b baseContainer) DstEl() interface{} {
	elementValue := reflect.New(b.elementType)
	return elementValue.Interface()
}

func (b baseContainer) dstElementToValue(dstEl interface{}) reflect.Value {
	val := reflect.ValueOf(dstEl)
	if !b.isElementAPointer {
		val = reflect.Indirect(val)
	}
	return val
}

type mapContainer struct {
	*baseContainer
}

func (m mapContainer) AddElementWithSubkey(key, subkey string, value interface{}) {
	if subkey != "" {
		key += "-" + subkey
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

type sliceContainer struct {
	*baseContainer
}

func (s sliceContainer) AddElementWithSubkey(key, _ string, value interface{}) {
	s.AddElement(key, value)
}

func (s sliceContainer) AddElement(_ string, value interface{}) {
	s.cntValue = reflect.Append(s.cntValue, s.dstElementToValue(value))
	s.assignableValue.Set(s.cntValue)
}

func (s sliceContainer) InitWithSize(size int) {
	if s.cntValue.IsNil() {
		s.cntValue.Set(reflect.MakeSlice(s.cntType, 0, size))
	}
}

func NewContainer(dst interface{}) (Container, error) {
	reflectValue := reflect.Indirect(reflect.ValueOf(dst))
	var result Container
	base := &baseContainer{
		assignableValue: reflectValue,
	}
	// the check is needed if dst is created via a function which returns an interface{}
	if _, ok := dst.(*interface{}); ok {
		reflectValue = reflectValue.Elem()
	}
	kind := reflectValue.Kind()
	base.cntValue = reflectValue

	switch kind {
	case reflect.Map:
		mapType := reflectValue.Type()
		// get the type of the key.
		keyType := mapType.Key()
		if keyType.Kind() != reflect.String {
			return nil, errors.Errorf("dst key type must be a string, %v given", keyType.Kind())
		}
		base.cntType = mapType
		result = mapContainer{baseContainer: base}
	case reflect.Slice:
		base.cntType = reflectValue.Type()
		result = sliceContainer{baseContainer: base}
	default:
		return nil, errors.Wrapf(ErrNonContainerType, "dst must be a map or a slice instead of %v", reflectValue.Type())
	}
	base.elementType = base.cntType.Elem()
	if base.elementType.Kind() == reflect.Ptr {
		// get the dst that the pointer elementType points to.
		base.elementType = base.elementType.Elem()
		base.isElementAPointer = true
	}
	return result, nil
}

var _ Container = mapContainer{}
var _ Container = sliceContainer{}

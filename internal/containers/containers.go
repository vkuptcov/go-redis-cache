package containers

import (
	"reflect"

	"github.com/pkg/errors"
)

var (
	ErrNonContainerType = errors.New("dst must be a map or a slice")
)

type containerInt interface {
	DstEl() interface{}
	AddElement(key string, value interface{})
	InitWithSize(size int)
}

type baseContainer struct {
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
		val = val.Elem()
	}
	return val
}

type mapContainer struct {
	*baseContainer
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

func (s sliceContainer) AddElement(_ string, value interface{}) {
	s.cntValue.Set(reflect.Append(s.cntValue, s.dstElementToValue(value)))
}

func (s sliceContainer) InitWithSize(size int) {
	if s.cntValue.IsNil() {
		s.cntValue.Set(reflect.MakeSlice(s.cntType, 0, size))
	}
}

func NewContainer(dst interface{}) (containerInt, error) {
	reflectValue := reflect.ValueOf(dst)
	if reflectValue.Kind() == reflect.Ptr {
		// get the dst that the pointer reflectValue points to.
		reflectValue = reflectValue.Elem()
	}

	var result containerInt
	base := &baseContainer{
		cntValue: reflectValue,
	}
	switch reflectValue.Kind() {
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

var _ containerInt = mapContainer{}
var _ containerInt = sliceContainer{}

package containers

import (
	"reflect"

	"github.com/pkg/errors"
)

type Container interface {
	DstEl() interface{}
	AddElementWithSubkey(key, subkey string, value interface{})
	AddElement(key string, value interface{})
	InitWithSize(size int)
	IsMultiElementContainer() bool
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

func (b baseContainer) IsMultiElementContainer() bool {
	return true
}

func (b baseContainer) dstElementToValue(dstEl interface{}) reflect.Value {
	val := reflect.ValueOf(dstEl)
	if !b.isElementAPointer {
		val = reflect.Indirect(val)
	}
	return val
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
		return singleElement{
			assignableValue: base.assignableValue,
			dst:             dst,
		}, nil
	}
	base.elementType = base.cntType.Elem()
	if base.elementType.Kind() == reflect.Ptr {
		// get the dst that the pointer elementType points to.
		base.elementType = base.elementType.Elem()
		base.isElementAPointer = true
	}
	return result, nil
}

package internal

import (
	"reflect"

	"github.com/pkg/errors"
)

func newDataTransformer(data interface{}, itemToCacheKeyFn func(it interface{}) (key, field string)) interface {
	getItems() ([]*Item, error)
} {
	v := reflect.ValueOf(data)
	switch kind := v.Kind(); kind {
	case reflect.Map:
		return mapTransformer{v}
	case reflect.Slice:
		return sliceTransformer{
			v:           v,
			itemToKeyFn: itemToCacheKeyFn,
		}
	default:
		// @todo support single element here
		panic(errors.Wrapf(ErrWrongLoadFnType, "Unsupported kind %q", kind))
	}
}

type mapTransformer struct {
	v reflect.Value
}

func (mt mapTransformer) getItems() ([]*Item, error) {
	v := mt.v
	if v.Len() == 0 {
		return nil, nil
	}
	mapType := v.Type()
	keyType := mapType.Key()
	if keyType.Kind() != reflect.String {
		return nil, errors.Errorf("dst key type must be a string, %v given", keyType.Kind())
	}
	iter := v.MapRange()
	items := make([]*Item, 0, v.Len())
	for iter.Next() {
		val := iter.Value().Interface()
		if item, ok := val.(*Item); ok {
			// @todo add possibility to use the key from the map
			items = append(items, item)
		} else {
			key := iter.Key().String()
			items = append(items, &Item{
				Key:   key,
				Value: val,
			})
		}
	}
	return items, nil
}

type sliceTransformer struct {
	v           reflect.Value
	itemToKeyFn func(it interface{}) (key, field string)
}

func (st sliceTransformer) getItems() ([]*Item, error) {
	v := st.v
	if v.Len() == 0 {
		return nil, nil
	}
	items := make([]*Item, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		val := v.Index(i).Interface()
		if item, ok := val.(*Item); ok {
			items = append(items, item)
		} else {
			key, field := st.itemToKeyFn(val)
			items = append(items, &Item{
				Key:   key,
				Field: field,
				Value: val,
			})
		}
	}
	return items, nil
}

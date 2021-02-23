package internal

import (
	"reflect"

	"github.com/pkg/errors"
)

func newDataTransformer(absentKeys []string, data interface{}, itemToCacheKeyFn func(it interface{}) (key, field string)) (interface {
	getItems() ([]*Item, error)
},
	error,
) {
	v := reflect.ValueOf(data)
	switch kind := v.Kind(); kind {
	case reflect.Map:
		return mapTransformer{v}, nil
	case reflect.Slice:
		return sliceTransformer{
			v:           v,
			itemToKeyFn: itemToCacheKeyFn,
		}, nil
	default:
		if len(absentKeys) != 1 {
			return nil, errors.Wrapf(ErrWrongLoadFnType, "Unsupported kind %q with %d keys", kind, len(absentKeys))
		}
		return singleElementTransformer{
			key:  absentKeys[0],
			data: data,
		}, nil
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

type singleElementTransformer struct {
	key  string
	data interface{}
}

func (st singleElementTransformer) getItems() ([]*Item, error) {
	if item, ok := st.data.(*Item); ok {
		return []*Item{item}, nil
	}
	return []*Item{
		{
			Key:   st.key,
			Value: st.data,
		},
	}, nil
}

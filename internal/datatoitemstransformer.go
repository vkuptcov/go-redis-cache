package internal

import (
	"reflect"

	"github.com/pkg/errors"

	"github.com/vkuptcov/go-redis-cache/v8/cachekeys"
)

func newDataTransformer(absentKeys []string, data interface{}, itemToCacheKeyFn func(it interface{}) (key, field string)) interface {
	getItems() ([]*Item, error)
} {
	v := reflect.ValueOf(data)
	switch kind := v.Kind(); kind {
	case reflect.Map:
		return mapTransformer{
			v:           v,
			itemToKeyFn: itemToCacheKeyFn,
		}
	case reflect.Slice:
		return sliceTransformer{
			v:           v,
			itemToKeyFn: itemToCacheKeyFn,
		}
	default:
		return singleElementTransformer{
			keys:        absentKeys,
			data:        data,
			itemToKeyFn: itemToCacheKeyFn,
		}
	}
}

type mapTransformer struct {
	v           reflect.Value
	itemToKeyFn func(it interface{}) (key, field string)
}

func (mt mapTransformer) getItems() ([]*Item, error) {
	v := mt.v
	if v.Len() == 0 {
		return nil, nil
	}
	mapType := v.Type()
	keyType := mapType.Key()
	if keyType.Kind() != reflect.String {
		return nil, errors.Wrapf(ErrNonStringKey, "dst key type must be a string, %v given", keyType.Kind())
	}
	iter := v.MapRange()
	items := make([]*Item, 0, v.Len())
	for iter.Next() {
		val := iter.Value().Interface()
		if item, ok := val.(*Item); ok {
			// @todo add possibility to use the key from the map
			items = append(items, item)
		} else {
			var key, field string
			if mt.itemToKeyFn != nil {
				key, field = mt.itemToKeyFn(val)
			} else {
				mapKey := iter.Key().String()
				key, field = cachekeys.SplitKeyAndField(mapKey)
			}
			items = append(items, &Item{
				Key:   key,
				Field: field,
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
			if st.itemToKeyFn == nil {
				return items, errors.WithStack(ErrItemToCacheKeyFnRequired)
			}
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
	keys        []string
	data        interface{}
	itemToKeyFn func(it interface{}) (key, field string)
}

func (st singleElementTransformer) getItems() ([]*Item, error) {
	if len(st.keys) != 1 {
		return nil, errors.Wrapf(ErrWrongLoadFnType, "Unsupported loaded element %T with %d keys", st.data, len(st.keys))
	}
	if item, ok := st.data.(*Item); ok {
		return []*Item{item}, nil
	}
	var key, field string
	if st.itemToKeyFn != nil {
		key, field = st.itemToKeyFn(st.data)
	} else {
		key, field = cachekeys.SplitKeyAndField(st.keys[0])
	}
	return []*Item{
		{
			Key:   key,
			Field: field,
			Value: st.data,
		},
	}, nil
}

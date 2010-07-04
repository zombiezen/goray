//
//	yaml/data/values.go
//	goray
//
//	Created by Ross Light on 2010-07-04.
//

package data

import (
	"reflect"
)

func AsBool(data interface{}) (b bool, ok bool) {
	b, ok = data.(bool)
	return
}

func AsFloat(data interface{}) (f float64, ok bool) {
	val := reflect.NewValue(data)
	ok = true

	switch realVal := val.(type) {
	case *reflect.FloatValue:
		f = realVal.Get()
	case *reflect.IntValue:
		f = float64(realVal.Get())
	case *reflect.UintValue:
		f = float64(realVal.Get())
	default:
		ok = false
	}
	return
}

func AsUint(data interface{}) (i uint64, ok bool) {
	val := reflect.NewValue(data)
	ok = true

	switch realVal := val.(type) {
	case *reflect.UintValue:
		i = realVal.Get()
	case *reflect.IntValue:
		if realVal.Get() >= 0 {
			i = uint64(realVal.Get())
		} else {
			ok = false
		}
	default:
		ok = false
	}
	return
}

func AsInt(data interface{}) (i int64, ok bool) {
	val := reflect.NewValue(data)
	ok = true

	switch realVal := val.(type) {
	case *reflect.IntValue:
		i = realVal.Get()
	case *reflect.UintValue:
		i = int64(realVal.Get())
	default:
		ok = false
	}
	return
}

func AsSequence(data interface{}) (seq []interface{}, ok bool) {
	seq, ok = data.([]interface{})
	return
}

func AsMap(data interface{}) (m map[interface{}]interface{}, ok bool) {
	m, ok = data.(map[interface{}]interface{})
	return
}

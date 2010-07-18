//
//	yaml/data/values.go
//	goray
//
//	Created by Ross Light on 2010-07-04.
//

/*
	The data package has various ways to work with generic data.  It was
	designed for YAML data, but many of the same functions are useful for
	generic mappings and such.
*/
package data

import (
	"reflect"
)

// Sequence is an ordered collection of values.
type Sequence []interface{}

// Map is an unordered collection of values associated with key values.
type Map map[interface{}]interface{}

// AsBool converts an untyped value to a boolean.
func AsBool(data interface{}) (b bool, ok bool) {
	b, ok = data.(bool)
	return
}

// AsFloat converts an untyped value to a floating-point number.
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

// AsUint converts an untyped value to an unsigned integer.
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

// AsInt converts an untyped value to a signed integer.
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

// AsSequence converts an untyped value to a sequence of values.
func AsSequence(data interface{}) (seq Sequence, ok bool) {
	seq, ok = data.([]interface{})
	return
}

// AsMap converts an untyped value to a map of values.
func AsMap(data interface{}) (m Map, ok bool) {
	m, ok = data.(map[interface{}]interface{})
	return
}

// HasKeys returns whether a given map contains all of the keys given.
func (m Map) HasKeys(keys ...interface{}) bool {
	for _, k := range keys {
		if _, found := m[k]; !found {
			return false
		}
	}
	return true
}

// CopyMap creates a shallow copy of a map.
func (m Map) Copy() (clone Map) {
	clone = make(Map, len(m))
	for k, v := range m {
		clone[k] = v
	}
	return
}

// SetDefault adds a new key to a map if the key isn't already present, and
// returns the latest value for the key.
func (m Map) SetDefault(k, d interface{}) (v interface{}) {
	v, ok := m[k]
	if !ok {
		m[k] = d
		v = d
	}
	return
}

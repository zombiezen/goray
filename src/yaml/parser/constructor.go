//
//	yaml/parser/constructor.go
//	goray
//
//	Created by Ross Light on 2010-07-04.
//

package parser

import (
	"math"
	"os"
	"strconv"
)

func init() {
	m := make(ConstructorMap)
	m[StringTag] = ConstructorFunc(constructStr)
	m[SequenceTag] = ConstructorFunc(constructSeq)
	m[MappingTag] = ConstructorFunc(constructMap)
	m[NullTag] = ConstructorFunc(constructNull)
	m[BoolTag] = ConstructorFunc(constructBool)
	m[IntTag] = ConstructorFunc(constructInt)
	m[FloatTag] = ConstructorFunc(constructFloat)
	DefaultConstructor = m
}

type Constructor interface {
	Construct(Node) (data interface{}, err os.Error)
}

type ConstructorFunc func(Node) (interface{}, os.Error)

func (f ConstructorFunc) Construct(n Node) (interface{}, os.Error) { return f(n) }

type ConstructorMap map[string]Constructor

func (m ConstructorMap) Construct(n Node) (data interface{}, err os.Error) {
	if c, ok := m[n.Tag()]; ok {
		return c.Construct(n)
	}
	err = os.NewError("Constructor has no rule for " + n.Tag())
	return
}

var DefaultConstructor Constructor

func constructStr(n Node) (data interface{}, err os.Error) {
	node, ok := n.(*Scalar)
	if !ok {
		err = os.NewError("Non-scalar given to string")
		return
	}
	data = node.String()
	return
}

func constructSeq(n Node) (data interface{}, err os.Error) {
	node, ok := n.(*Sequence)
	if !ok {
		err = os.NewError("Non-sequence given to sequence")
		return
	}

	a := make([]interface{}, node.Len())
	for i := 0; i < node.Len(); i++ {
		a[i] = node.At(i).Data()
	}
	data = a
	return
}

func constructMap(n Node) (data interface{}, err os.Error) {
	node, ok := n.(*Mapping)
	if !ok {
		err = os.NewError("Non-mapping given to map")
		return
	}

	m := make(map[interface{}]interface{}, node.Len())
	for _, pair := range node.Pairs {
		k := pair.Key.Data()
		v := pair.Value.Data()
		m[k] = v
	}
	data = m
	return
}

func constructNull(n Node) (data interface{}, err os.Error) {
	_, isScalar := n.(*Scalar)
	_, isEmpty := n.(*Empty)
	if !isScalar || !isEmpty {
		err = os.NewError("Non-scalar tagged as null")
		return
	}
	return nil, nil
}

func constructBool(n Node) (data interface{}, err os.Error) {
	var s string

	if scalar, ok := n.(*Scalar); ok {
		s = scalar.Value
	} else {
		err = os.NewError("Non-scalar tagged as bool")
		return
	}
	groups := csBoolPat.MatchStrings(s)

	if len(groups) > 0 {
		if groups[1] != "" {
			return true, nil
		} else if groups[2] != "" {
			return false, nil
		}
	}

	err = os.NewError("Value is an invalid boolean: " + s)
	return
}

func constructInt(n Node) (data interface{}, err os.Error) {
	var s string

	if scalar, ok := n.(*Scalar); ok {
		s = scalar.Value
	} else {
		err = os.NewError("Non-scalar tagged as int")
		return
	}

	if csDecPat.MatchString(s) {
		return strconv.Atoi64(s)
	} else if groups := csHexPat.MatchStrings(s); len(groups) > 0 {
		return strconv.Btoui64(groups[1], 16)
	} else if groups := csOctPat.MatchStrings(s); len(groups) > 0 {
		return strconv.Btoui64(groups[1], 8)
	}

	err = os.NewError("Value is an invalid int: " + s)
	return
}

func constructFloat(n Node) (data interface{}, err os.Error) {
	var s string

	if scalar, ok := n.(*Scalar); ok {
		s = scalar.Value
	} else {
		err = os.NewError("Non-scalar tagged as float")
		return
	}

	switch {
	case csFloatPat.MatchString(s):
		return strconv.Atof64(s)
	case csInfPat.MatchString(s):
		sign := 1
		if s[0] == '-' {
			sign = -1
		}
		return math.Inf(sign), nil
	case csNanPat.MatchString(s):
		return math.NaN(), nil
	}

	err = os.NewError("Value is an invalid float: " + s)
	return
}

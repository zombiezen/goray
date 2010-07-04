//
//	yaml/data/constructor.go
//	goray
//
//	Created by Ross Light on 2010-07-04.
//

package data

import (
	"math"
	"os"
	"strconv"
	"yaml/parser"
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

type Constructor parser.Constructor

type ConstructorFunc func(parser.Node) (interface{}, os.Error)

func (f ConstructorFunc) Construct(n parser.Node) (interface{}, os.Error) { return f(n) }

type ConstructorMap map[string]Constructor

func (m ConstructorMap) Construct(n parser.Node) (data interface{}, err os.Error) {
	if c, ok := m[n.Tag()]; ok {
		return c.Construct(n)
	}
	err = os.NewError("Constructor has no rule for " + n.Tag())
	return
}

var DefaultConstructor Constructor

func constructStr(n parser.Node) (data interface{}, err os.Error) {
	node, ok := n.(*parser.Scalar)
	if !ok {
		err = os.NewError("Non-scalar given to string")
		return
	}
	data = node.String()
	return
}

func constructSeq(n parser.Node) (data interface{}, err os.Error) {
	node, ok := n.(*parser.Sequence)
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

func constructMap(n parser.Node) (data interface{}, err os.Error) {
	node, ok := n.(*parser.Mapping)
	if !ok {
		err = os.NewError("Non-mapping given to map")
		return
	}

	m := make(map[interface{}]interface{}, node.Len())
	for _, pair := range node.Pairs {
		m[pair.Key.Data()] = pair.Value.Data()
	}
	data = m
	return
}

func constructNull(n parser.Node) (data interface{}, err os.Error) {
	_, isScalar := n.(*parser.Scalar)
	_, isEmpty := n.(*parser.Empty)
	if !isScalar && !isEmpty {
		err = os.NewError("Non-scalar tagged as null")
		return
	}
	return nil, nil
}

func constructBool(n parser.Node) (data interface{}, err os.Error) {
	var s string

	if scalar, ok := n.(*parser.Scalar); ok {
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

func constructInt(n parser.Node) (data interface{}, err os.Error) {
	var s string

	if scalar, ok := n.(*parser.Scalar); ok {
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

func constructFloat(n parser.Node) (data interface{}, err os.Error) {
	var s string

	if scalar, ok := n.(*parser.Scalar); ok {
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

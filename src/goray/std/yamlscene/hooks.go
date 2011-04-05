//
//	goray/std/yamlscene/hooks.go
//	goray
//
//	Created by Ross Light on 2010-07-04.
//

package yamlscene

import (
	"fmt"
	"os"
	"goray/core/color"
	"goray/core/vector"

	yamldata "goyaml.googlecode.com/hg/data"
	"goyaml.googlecode.com/hg/parser"
)

type ConstructError struct {
	os.Error
	Node parser.Node
}

func (err ConstructError) String() string {
	return fmt.Sprintf("line %d: %s", err.Node.Start().Line, err.Error)
}

type MapConstruct func(yamldata.Map) (interface{}, os.Error)

func (f MapConstruct) Construct(n parser.Node) (data interface{}, err os.Error) {
	if node, ok := n.(*parser.Mapping); ok {
		data, err = f(node.Map())
	} else {
		err = os.NewError("Constructor requires a mapping")
	}

	if err != nil {
		err = ConstructError{err, n}
	}
	return
}

var Constructor yamldata.ConstructorMap = yamldata.ConstructorMap{
	Prefix + "rgb":  yamldata.ConstructorFunc(constructRGB),
	Prefix + "rgba": yamldata.ConstructorFunc(constructRGBA),
	Prefix + "vec":  yamldata.ConstructorFunc(constructVector),
}

func float64Sequence(n parser.Node) (data []float64, ok bool) {
	seq, ok := n.(*parser.Sequence)
	if !ok {
		return
	}

	data = make([]float64, seq.Len())
	for i := 0; i < seq.Len(); i++ {
		data[i], ok = yamldata.AsFloat(seq.At(i).Data())
		if !ok {
			return
		}
	}
	return
}

func floatSequence(n parser.Node) (data []float64, ok bool) {
	f64Data, ok := float64Sequence(n)
	if ok {
		data = make([]float64, len(f64Data))
		for i, f := range f64Data {
			data[i] = float64(f)
		}
	}
	return
}

func constructRGB(n parser.Node) (data interface{}, err os.Error) {
	comps, ok := floatSequence(n)
	if !ok || len(comps) != 3 {
		err = os.NewError("RGB must be a sequence of 3 floats")
		return
	}
	return color.RGB{comps[0], comps[1], comps[2]}, nil
}

func constructRGBA(n parser.Node) (data interface{}, err os.Error) {
	comps, ok := floatSequence(n)
	if !ok || len(comps) != 4 {
		err = os.NewError("RGBA must be a sequence of 4 floats")
		return
	}
	return color.RGBA{comps[0], comps[1], comps[2], comps[3]}, nil
}

func constructVector(n parser.Node) (data interface{}, err os.Error) {
	comps, ok := float64Sequence(n)
	if !ok || len(comps) != 3 {
		err = os.NewError("Vector must be a sequence of 3 floats")
		return
	}
	return vector.Vector3D{comps[0], comps[1], comps[2]}, nil
}

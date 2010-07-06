//
//	goray/std/yamlscene/hooks.go
//	goray
//
//	Created by Ross Light on 2010-07-04.
//

package yamlscene

import (
	"os"
	"goray/core/color"
	"goray/core/vector"

	orthocam "goray/std/cameras/ortho"
	directlight "goray/std/integrators/directlight"
	pointlight "goray/std/lights/point"
	debugmaterial "goray/std/materials/debug"
	mesh "goray/std/objects/mesh"

	yamldata "yaml/data"
	"yaml/parser"
)

var Constructor yamldata.ConstructorMap = yamldata.ConstructorMap{
	Prefix + "rgb":  yamldata.ConstructorFunc(constructRGB),
	Prefix + "rgba": yamldata.ConstructorFunc(constructRGBA),
	Prefix + "vec":  yamldata.ConstructorFunc(constructVector),

	StdPrefix + "cameras/ortho":           yamldata.ConstructorFunc(orthocam.Construct),
	StdPrefix + "integrators/directlight": yamldata.ConstructorFunc(directlight.Construct),
	StdPrefix + "lights/point":            yamldata.ConstructorFunc(pointlight.Construct),
	StdPrefix + "materials/debug":         yamldata.ConstructorFunc(debugmaterial.Construct),
	StdPrefix + "objects/mesh":            yamldata.ConstructorFunc(mesh.Construct),
}

func floatSequence(n parser.Node) (data []float, ok bool) {
	seq, ok := n.(*parser.Sequence)
	if !ok {
		return
	}

	data = make([]float, seq.Len())
	for i := 0; i < seq.Len(); i++ {
		var f float64
		f, ok = yamldata.AsFloat(seq.At(i).Data())
		if !ok {
			return
		}
		data[i] = float(f)
	}
	return
}

func constructRGB(n parser.Node) (data interface{}, err os.Error) {
	comps, ok := floatSequence(n)
	if !ok || len(comps) != 3 {
		err = os.NewError("RGB must be a sequence of 3 floats")
		return
	}
	return color.NewRGB(comps[0], comps[1], comps[2]), nil
}

func constructRGBA(n parser.Node) (data interface{}, err os.Error) {
	comps, ok := floatSequence(n)
	if !ok || len(comps) != 4 {
		err = os.NewError("RGBA must be a sequence of 4 floats")
		return
	}
	return color.NewRGBA(comps[0], comps[1], comps[2], comps[3]), nil
}

func constructVector(n parser.Node) (data interface{}, err os.Error) {
	comps, ok := floatSequence(n)
	if !ok || len(comps) != 3 {
		err = os.NewError("Vector must be a sequence of 3 floats")
		return
	}
	return vector.New(comps[0], comps[1], comps[2]), nil
}

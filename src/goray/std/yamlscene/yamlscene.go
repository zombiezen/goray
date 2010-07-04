//
//	goray/std/yamlscene/yamlscene.go
//	goray
//
//	Created by Ross Light on 2010-07-03.
//

package yamlscene

import (
	"io"
	"os"
	"goray/core/color"
	"goray/core/integrator"
	"goray/core/scene"
	yamlData "yaml/data"
	"yaml/parser"
)

const (
	Prefix    = "tag:goray/"
	StdPrefix = Prefix + "std/"
)

var Constructor yamlData.ConstructorMap

func init() {
	Constructor = make(yamlData.ConstructorMap)
	Constructor[Prefix+"rgb"] = yamlData.ConstructorFunc(constructRGB)
	Constructor[Prefix+"rgba"] = yamlData.ConstructorFunc(constructRGBA)
}

func Load(r io.Reader, sc *scene.Scene) (i integrator.Integrator, err os.Error) {
	// Parse
	p := parser.New(r, yamlData.CoreSchema, yamlData.ConstructorFunc(realConstructor))
	doc, err := p.ParseDocument()
	if err != nil {
		return
	}

	// Set up scene!
	root := doc.Content.Data().(map[interface{}]interface{})
	i = root["integrator"].(integrator.Integrator)
	return
}

func realConstructor(n parser.Node) (interface{}, os.Error) {
	if _, ok := Constructor[n.Tag()]; ok {
		return Constructor.Construct(n)
	}
	return yamlData.DefaultConstructor.Construct(n)
}

func floatSequence(n parser.Node) (data []float, ok bool) {
	seq, ok := n.(*parser.Sequence)
	if !ok {
		return
	}
	
	data = make([]float, seq.Len())
	for i := 0; i < seq.Len(); i++ {
		var f float64
		f, ok = yamlData.AsFloat(seq.At(i).Data())
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

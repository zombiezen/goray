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
	"yaml/parser"
)

const (
	Prefix = "tag:goray/"
	StdPrefix = Prefix + "std/"
)

var Constructor parser.ConstructorMap

func init() {
	Constructor = make(parser.ConstructorMap)
	Constructor[Prefix + "rgb"] = parser.ConstructorFunc(constructRGB)
	Constructor[Prefix + "rgba"] = parser.ConstructorFunc(constructRGBA)
}

func Load(r io.Reader, sc *scene.Scene) (i integrator.Integrator, err os.Error) {
	// Parse
	p := parser.New(r, nil, parser.ConstructorFunc(realConstructor))
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
	return parser.DefaultConstructor.Construct(n)
}

func constructRGB(n parser.Node) (data interface{}, err os.Error) {
	if node, ok := n.(*parser.Sequence); ok {
		if node.Len() == 3 {
			rComp, rOk := parser.GetNodeFloat(node.At(0))
			gComp, gOk := parser.GetNodeFloat(node.At(1))
			bComp, bOk := parser.GetNodeFloat(node.At(2))
			if rOk && gOk && bOk {
				data = color.NewRGB(rComp, gComp, bComp)
			} else {
				err = os.NewError("RGB must have 3 floats")
			}
		} else {
			err = os.NewError("RGB must have 3 components")
		}
	} else {
		err = os.NewError("RGB must be a sequence")
	}
	
	return
}

func constructRGBA(n parser.Node) (data interface{}, err os.Error) {
	if node, ok := n.(*parser.Sequence); ok {
		if node.Len() == 4 {
			rComp, rOk := parser.GetNodeFloat(node.At(0))
			gComp, gOk := parser.GetNodeFloat(node.At(1))
			bComp, bOk := parser.GetNodeFloat(node.At(2))
			aComp, aOk := parser.GetNodeFloat(node.At(3))
			
			if rOk && gOk && bOk && aOk {
				data = color.NewRGBA(rComp, gComp, bComp, aComp)
			} else {
				err = os.NewError("RGBA must have 4 floats")
			}
		} else {
			err = os.NewError("RGBA must have 4 components")
		}
	} else {
		err = os.NewError("RGBA must be a sequence")
	}
	
	return
}

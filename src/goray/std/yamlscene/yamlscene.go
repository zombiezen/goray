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
	"goray/core/integrator"
	"goray/core/scene"
	yamlData "yaml/data"
	"yaml/parser"
)

const (
	Prefix    = "tag:goray/"
	StdPrefix = Prefix + "std/"
)

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

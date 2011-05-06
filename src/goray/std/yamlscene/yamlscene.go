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
	"goray/core/camera"
	"goray/core/integrator"
	"goray/core/light"
	"goray/core/object"
	"goray/core/scene"
	yamldata "goyaml.googlecode.com/hg/data"
	"goyaml.googlecode.com/hg/parser"
)

const (
	Prefix    = "tag:goray/"
	StdPrefix = Prefix + "std/"
)

type Params map[string]interface{}

func Load(r io.Reader, sc *scene.Scene, params Params) (i integrator.Integrator, err os.Error) {
	// Parse
	p := parser.New(r, yamldata.CoreSchema, yamldata.ConstructorFunc(realConstructor), params)
	doc, err := p.ParseDocument()
	if err != nil {
		return
	}

	// Set up scene!
	root := doc.Content.(*parser.Mapping).Map()

	objects, _ := yamldata.AsSequence(root["objects"])
	for i, _ := range objects {
		obj := objects[i].(object.Object3D)
		sc.AddObject(obj)
	}

	lights, _ := yamldata.AsSequence(root["lights"])
	for i, _ := range lights {
		l := lights[i].(light.Light)
		sc.AddLight(l)
	}

	camera, _ := root["camera"].(camera.Camera)
	sc.SetCamera(camera)

	// Get integrator and finish
	i = root["integrator"].(integrator.Integrator)
	return
}

func realConstructor(n parser.Node, userData interface{}) (interface{}, os.Error) {
	if _, ok := Constructor[n.Tag()]; ok {
		return Constructor.Construct(n, userData)
	}
	return yamldata.DefaultConstructor.Construct(n, userData)
}

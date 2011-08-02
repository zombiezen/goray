/*
	Copyright (c) 2011 Ross Light.
	Copyright (c) 2005 Mathias Wein, Alejandro Conty, and Alfredo de Greef.

	This file is part of goray.

	goray is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	goray is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with goray.  If not, see <http://www.gnu.org/licenses/>.
*/

package yamlscene

import (
	"io"
	"os"

	"goray"
	yamldata "goyaml.googlecode.com/hg/data"
	"goyaml.googlecode.com/hg/parser"
)

const (
	Prefix    = "tag:goray/"
	StdPrefix = Prefix + "std/"
)

type Params map[string]interface{}

func Load(r io.Reader, sc *goray.Scene, params Params) (i goray.Integrator, err os.Error) {
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
		obj := objects[i].(goray.Object3D)
		sc.AddObject(obj)
	}

	lights, _ := yamldata.AsSequence(root["lights"])
	for i, _ := range lights {
		l := lights[i].(goray.Light)
		sc.AddLight(l)
	}

	camera, _ := root["camera"].(goray.Camera)
	sc.SetCamera(camera)

	// Get integrator and finish
	i = root["integrator"].(goray.Integrator)
	return
}

func realConstructor(n parser.Node, userData interface{}) (interface{}, os.Error) {
	if _, ok := Constructor[n.Tag()]; ok {
		return Constructor.Construct(n, userData)
	}
	return yamldata.DefaultConstructor.Construct(n, userData)
}

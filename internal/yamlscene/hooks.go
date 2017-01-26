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
	"errors"

	"bitbucket.org/zombiezen/math3/vec64"
	"zombiezen.com/go/goray/internal/color"
	"zombiezen.com/go/goray/internal/goray"
	yamldata "zombiezen.com/go/goray/internal/yaml/data"
	"zombiezen.com/go/goray/internal/yaml/parser"
)

type MapConstruct func(yamldata.Map) (interface{}, error)

func (f MapConstruct) Construct(n parser.Node, userData interface{}) (data interface{}, err error) {
	if node, ok := n.(*parser.Mapping); ok {
		data, err = f(node.Map())
	} else {
		err = errors.New("Constructor requires a mapping")
	}
	return
}

var Constructor yamldata.ConstructorMap = yamldata.ConstructorMap{
	Prefix + "rgb":  yamldata.ConstructorFunc(constructRGB),
	Prefix + "rgba": yamldata.ConstructorFunc(constructRGBA),
	Prefix + "vec":  yamldata.ConstructorFunc(constructVector),

	StdPrefix + "objects/mesh": MapConstruct(constructMesh),
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

func constructRGB(n parser.Node, userData interface{}) (data interface{}, err error) {
	comps, ok := floatSequence(n)
	if !ok || len(comps) != 3 {
		err = errors.New("RGB must be a sequence of 3 floats")
		return
	}
	return color.RGB{comps[0], comps[1], comps[2]}, nil
}

func constructRGBA(n parser.Node, userData interface{}) (data interface{}, err error) {
	comps, ok := floatSequence(n)
	if !ok || len(comps) != 4 {
		err = errors.New("RGBA must be a sequence of 4 floats")
		return
	}
	return color.RGBA{comps[0], comps[1], comps[2], comps[3]}, nil
}

func constructVector(n parser.Node, userData interface{}) (data interface{}, err error) {
	comps, ok := float64Sequence(n)
	if !ok || len(comps) != 3 {
		err = errors.New("Vector must be a sequence of 3 floats")
		return
	}
	return vec64.Vector{comps[0], comps[1], comps[2]}, nil
}

func constructMesh(m yamldata.Map) (data interface{}, err error) {
	m = m.Copy()
	m.SetDefault("vertices", []interface{}{})
	m.SetDefault("uvs", []interface{}{})
	m.SetDefault("faces", []interface{}{})

	vertices, ok := yamldata.AsSequence(m["vertices"])
	if !ok {
		err = errors.New("Vertices must be a sequence")
		return
	}

	uvs, ok := yamldata.AsSequence(m["uvs"])
	if !ok {
		err = errors.New("UVs must be a sequence")
		return
	}

	faces, ok := yamldata.AsSequence(m["faces"])
	if !ok {
		err = errors.New("Faces must be a sequence")
		return
	}

	mesh := goray.NewMesh(len(faces), false)

	var vertexData []vec64.Vector
	var uvData []goray.UV

	// Parse vertices
	// TODO: Error handling
	vertexData = make([]vec64.Vector, len(vertices))
	for i, _ := range vertices {
		vseq, _ := yamldata.AsSequence(vertices[i])
		x, _ := yamldata.AsFloat(vseq[0])
		y, _ := yamldata.AsFloat(vseq[1])
		z, _ := yamldata.AsFloat(vseq[2])
		vertexData[i] = vec64.Vector{x, y, z}
	}

	// Parse UVs
	// TODO: Error handling
	if len(uvs) > 0 {
		uvData = make([]goray.UV, len(uvs))
		for i, _ := range uvs {
			uvseq, _ := yamldata.AsSequence(uvs[i])
			u, _ := yamldata.AsFloat(uvseq[0])
			v, _ := yamldata.AsFloat(uvseq[1])
			uvData[i] = goray.UV{u, v}
		}
	}
	mesh.SetData(vertexData, nil, uvData)

	// Parse faces
	// TODO: Error handling
	for i, _ := range faces {
		fmap, _ := yamldata.AsMap(faces[i])
		// Vertices
		vindices, _ := yamldata.AsSequence(fmap["vertices"])
		va, _ := yamldata.AsInt(vindices[0])
		vb, _ := yamldata.AsInt(vindices[1])
		vc, _ := yamldata.AsInt(vindices[2])
		// UVs
		uva, uvb, uvc := -1, -1, -1
		if _, hasUVs := fmap["uvs"]; len(uvs) > 0 && hasUVs {
			uvindices, _ := yamldata.AsSequence(fmap["uvs"])
			uva, _ = yamldata.AsInt(uvindices[0])
			uvb, _ = yamldata.AsInt(uvindices[1])
			uvc, _ = yamldata.AsInt(uvindices[2])
		}
		// Create triangle
		tri := goray.NewTriangle(va, vb, vc, mesh)
		tri.SetUVs(uva, uvb, uvc)
		tri.SetMaterial(fmap["material"].(goray.Material))
		mesh.AddTriangle(tri)
	}

	return mesh, nil
}

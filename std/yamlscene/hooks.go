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

	"bitbucket.org/zombiezen/goray/color"
	"bitbucket.org/zombiezen/goray/vector"

	yamldata "bitbucket.org/zombiezen/goray/yaml/data"
	"bitbucket.org/zombiezen/goray/yaml/parser"
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
	return vector.Vector3D{comps[0], comps[1], comps[2]}, nil
}

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

// Package texmap provides a shader node that performs texture mapping with various options.
package texmap

import (
	"math"
	"os"

	"goray"
	"goray/matrix"
	"goray/shader"
	"goray/vector"
	"goray/std/yamlscene"

	yamldata "goyaml.googlecode.com/hg/data"
)

// Coordinates specifies which coordinate system to use during texture mapping.
type Coordinates int

const (
	UV        Coordinates = iota // UV-mapping
	Global                       // Global coordinates
	Orco                         // Original coordinates
	Transform                    // Transformation matrix
	Window                       // Viewport-relative
)

// A TextureMapper is a shader that applies a texture to geometry.
type TextureMapper struct {
	Texture          Texture         // The 2D/3D texture to apply
	Coordinates      Coordinates     // The coordinate system to use
	Projector        Projector       // The 3D projection type to use
	MapX, MapY, MapZ vector.Axis     // Axis re-mapping (use -1 to indicate zero)
	Transform        matrix.Matrix   // Transformation matrix (if using Transform coordinates)
	Scale, Offset    vector.Vector3D // Constant scale and offset for coordinates
	Scalar           bool            // Should the result be a scalar?
	BumpStrength     float64         // Bump mapping weight

	delta, deltaU, deltaV, deltaW float64
}

var _ shader.Node = &TextureMapper{}

// New creates a new texture mapper with the given parameters.
func New(tex Texture, coord Coordinates, scalar bool) (tmap *TextureMapper) {
	tmap = new(TextureMapper)
	tmap.Init(tex, coord, scalar)
	return
}

// Init initializes the mapper with default values.  You must call this method
// before using the mapper (but it is called automatically by New).
func (tmap *TextureMapper) Init(tex Texture, coord Coordinates, scalar bool) {
	const defaultDelta = 2.0e-4

	tmap.Texture = tex
	tmap.Coordinates = coord
	tmap.Projector = FlatMap
	tmap.MapX, tmap.MapY, tmap.MapZ = vector.X, vector.Y, vector.Z
	tmap.Transform = matrix.Identity
	tmap.Scale = vector.Vector3D{1.0, 1.0, 1.0}
	tmap.Offset = vector.Vector3D{}
	tmap.Scalar = scalar

	if discreteTex, ok := tex.(DiscreteTexture); ok {
		u, v, w := discreteTex.Resolution()
		tmap.deltaU = 1 / float64(u)
		tmap.deltaV = 1 / float64(v)
		if tex.Is3D() {
			tmap.deltaW = 1 / float64(w)
		} else {
			tmap.deltaW = 0
		}
		tmap.delta = math.Sqrt(tmap.deltaU*tmap.deltaU + tmap.deltaV*tmap.deltaV + tmap.deltaW*tmap.deltaW)
	} else {
		tmap.deltaU = defaultDelta
		tmap.deltaV = defaultDelta
		tmap.deltaW = defaultDelta
		tmap.delta = defaultDelta
	}
}

func (tmap *TextureMapper) textureCoordinates(state *goray.RenderState, sp goray.SurfacePoint) (p, n vector.Vector3D) {
	p, n = sp.Position, sp.GeometricNormal
	switch tmap.Coordinates {
	case UV:
		p = vector.Vector3D{sp.U, sp.V, 0}
	case Orco:
		p, n = sp.OrcoPosition, sp.OrcoNormal
	case Transform:
		p = matrix.VecMul(tmap.Transform, p)
	case Window:
		p = state.ScreenPos
	}
	return
}

func (tmap *TextureMapper) mapping(p, n vector.Vector3D) (texPt vector.Vector3D) {
	texPt = p
	if tmap.Coordinates == UV {
		// Normalize to [-1, 1]
		texPt = vector.Vector3D{2*texPt[vector.X] - 1, 2*texPt[vector.Y] - 1, texPt[vector.Z]}
	}

	// Map axes
	m := map[vector.Axis]float64{
		-1:       0.0,
		vector.X: texPt[vector.X],
		vector.Y: texPt[vector.Y],
		vector.Z: texPt[vector.Z],
	}
	texPt[vector.X] = m[tmap.MapX]
	texPt[vector.Y] = m[tmap.MapY]
	texPt[vector.Z] = m[tmap.MapZ]

	// Texture coordinate mapping
	texPt = tmap.Projector.Project(texPt, n)

	// Scale and offset
	texPt = vector.Add(vector.CompMul(texPt, tmap.Scale), tmap.Offset)
	return
}

func (tmap *TextureMapper) Eval(inputs []shader.Result, params shader.Params) (result shader.Result) {
	state := params["RenderState"].(*goray.RenderState)
	sp := params["SurfacePoint"].(goray.SurfacePoint)
	p := tmap.mapping(tmap.textureCoordinates(state, sp))

	// TODO: We may need to store both scalar and color.
	if tmap.Scalar {
		result = shader.Result{tmap.Texture.ScalarAt(p)}
	} else {
		col := tmap.Texture.ColorAt(p)
		result = shader.Result{col.Red(), col.Green(), col.Blue(), col.Alpha()}
	}
	return
}

func (tmap *TextureMapper) EvalDerivative(inputs []shader.Result, params shader.Params) (result shader.Result) {
	state := params["RenderState"].(*goray.RenderState)
	sp := params["SurfacePoint"].(goray.SurfacePoint)
	scale := tmap.Scale.Length()
	bstr := tmap.BumpStrength / scale
	if tmap.Coordinates == UV {
		var p1, p2 vector.Vector3D
		p1 = tmap.mapping(vector.Vector3D{sp.U - tmap.deltaU, sp.V, 0}, sp.GeometricNormal)
		p2 = tmap.mapping(vector.Vector3D{sp.U + tmap.deltaU, sp.V, 0}, sp.GeometricNormal)
		dfdu := (tmap.Texture.ScalarAt(p2) - tmap.Texture.ScalarAt(p1)) / tmap.deltaU
		p1 = tmap.mapping(vector.Vector3D{sp.U, sp.V - tmap.deltaV, 0}, sp.GeometricNormal)
		p2 = tmap.mapping(vector.Vector3D{sp.U, sp.V + tmap.deltaV, 0}, sp.GeometricNormal)
		dfdv := (tmap.Texture.ScalarAt(p2) - tmap.Texture.ScalarAt(p1)) / tmap.deltaV

		// Derivative is in UV-space, convert to shading space.
		vecU, vecV := sp.ShadingU, sp.ShadingV
		vecU[vector.Z], vecV[vector.Z] = dfdu, dfdv

		// Solve plane equation to get 1/0/df 0/1/df.
		norm := vector.Cross(vecU, vecV)
		if math.Fabs(norm[vector.Z]) > 1e-30 {
			nf := 1 / norm[vector.Z] * bstr
			result = shader.Result{norm[vector.X] * nf, norm[vector.Y] * nf}
		}
	} else {
		p, n := tmap.textureCoordinates(state, sp)
		delta := tmap.delta / scale
		du := vector.ScalarMul(sp.NormalU, delta)
		dv := vector.ScalarMul(sp.NormalV, delta)
		u1, u2 := tmap.mapping(vector.Sub(p, du), n), tmap.mapping(vector.Add(p, du), n)
		v1, v2 := tmap.mapping(vector.Sub(p, dv), n), tmap.mapping(vector.Add(p, dv), n)
		result = shader.Result{
			-bstr * (tmap.Texture.ScalarAt(u2) - tmap.Texture.ScalarAt(u1)) / delta,
			-bstr * (tmap.Texture.ScalarAt(v2) - tmap.Texture.ScalarAt(v1)) / delta,
		}
	}
	return
}

func (tmap *TextureMapper) ViewDependent() bool {
	// Texture mapping is view-independent. Window coordinates use render state.
	return false
}

func (tmap *TextureMapper) Dependencies() []shader.Node { return []shader.Node{} }

func init() {
	yamlscene.Constructor[yamlscene.StdPrefix+"shaders/texmap"] = yamlscene.MapConstruct(Construct)
}

func Construct(m yamldata.Map) (data interface{}, err os.Error) {
	// Defaults
	m = m.Copy()
	m.SetDefault("projection", "flat")
	m.SetDefault("mapAxes", yamldata.Sequence{"x", "y", "z"})
	m.SetDefault("scale", vector.Vector3D{1, 1, 1})
	m.SetDefault("offset", vector.Vector3D{})
	m.SetDefault("scalar", false)
	m.SetDefault("bumpStrength", 0.02)

	// Texture
	tex, ok := m["texture"].(Texture)
	if !ok {
		err = os.NewError("Texture mapper must be given a texture")
		return
	}

	// Coordinates
	var coord Coordinates
	coordString, ok := m["coordinates"].(string)
	if !ok {
		err = os.NewError("Texture mapper must have coordinates key")
		return
	}
	switch coordString {
	case "uv":
		coord = UV
	case "global":
		coord = Global
	case "orco":
		coord = Orco
	case "transform":
		coord = Transform
	case "window":
		coord = Window
	default:
		err = os.NewError("Unrecognized coordinate space: " + coordString)
		return
	}

	// Scalar
	scalar, ok := yamldata.AsBool(m["scalar"])
	if !ok {
		err = os.NewError("Scalar must be a boolean")
		return
	}
	tmap := New(tex, coord, scalar)

	// Projection
	projString, ok := m["projection"].(string)
	if !ok {
		err = os.NewError("projection must be string")
		return
	}
	switch projString {
	case "flat":
		tmap.Projector = FlatMap
	case "tube":
		tmap.Projector = TubeMap
	case "sphere":
		tmap.Projector = SphereMap
	case "cube":
		tmap.Projector = CubeMap
	default:
		err = os.NewError("Unrecognized projection: " + projString)
		return
	}

	// Axis mapping
	axisMap, ok := yamldata.AsSequence(m["mapAxes"])
	if !ok || len(axisMap) != 3 {
		err = os.NewError("mapAxes must be a 3-sequence")
		return
	}
	for i, axisItem := range axisMap {
		a, ok := axisItem.(string)
		// ^ heh.
		if !ok {
			err = os.NewError("Each item of mapAxes must be a string")
			return
		}
		switch i {
		case 0:
			tmap.MapX = constructMapAxis(a)
		case 1:
			tmap.MapY = constructMapAxis(a)
		case 2:
			tmap.MapZ = constructMapAxis(a)
		}
	}

	// Scale and offset
	tmap.Scale, ok = m["scale"].(vector.Vector3D)
	if !ok {
		err = os.NewError("scale must be a vector")
		return
	}
	tmap.Offset, ok = m["offset"].(vector.Vector3D)
	if !ok {
		err = os.NewError("offset must be a vector")
		return
	}

	// Bump Strength
	tmap.BumpStrength, ok = yamldata.AsFloat(m["bumpStrength"])
	if !ok {
		err = os.NewError("bumpStrength must be a number")
		return
	}

	// Finish
	return tmap, nil
}

func constructMapAxis(name string) (a vector.Axis) {
	switch name {
	case "x":
		a = vector.X
	case "y":
		a = vector.Y
	case "z":
		a = vector.Z
	case "none":
		a = -1
	default:
		a = -1
	}
	return
}

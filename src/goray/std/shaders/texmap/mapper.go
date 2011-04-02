//
//	goray/std/shaders/texmap/mapper.go
//	goray
//
//	Created by Ross Light on 2011-04-02.
//

package texmap

import (
	"goray/core/matrix"
	"goray/core/render"
	"goray/core/shader"
	"goray/core/surface"
	"goray/core/vector"
)

// Coordinates specifies which coordinate system to use during texture mapping.
type Coordinates int

// 
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
	MapX, MapY, MapZ vector.Axis     // Axis re-mapping (use -1 to indicate zero)
	Transform        matrix.Matrix   // Transformation matrix (if using Transform coordinates)
	Scale, Offset    vector.Vector3D // Constant scale and offset for coordinates
	Scalar           bool            // Should the result be a scalar?
}

var _ shader.Node = &TextureMapper{}

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
	// TODO: Texture coordinate mapping (tube, sphere, etc.)
	// Scale and offset
	texPt = vector.Add(vector.CompMul(texPt, tmap.Scale), tmap.Offset)
	return
}

func (tmap *TextureMapper) Eval(params map[string]interface{}, inputs []shader.Result) (result shader.Result) {
	sp := params["SurfacePoint"].(surface.Point)
	state := params["RenderState"].(*render.State)
	p, n := sp.Position, sp.GeometricNormal
	switch tmap.Coordinates {
	case UV:
		// TODO: eval_uv
	case Orco:
		p, n = sp.OrcoPosition, sp.OrcoNormal
	case Transform:
		p = matrix.VecMul(tmap.Transform, p)
	case Window:
		p = state.ScreenPos
	}
	p = tmap.mapping(p, n)
	// TODO: We may need to store both scalar and color.
	if tmap.Scalar {
		result = shader.Result{tmap.Texture.ScalarAt(p)}
	} else {
		col := tmap.Texture.ColorAt(p)
		result = shader.Result{col.Red(), col.Green(), col.Blue(), col.Alpha()}
	}
	return
}

func (tmap *TextureMapper) EvalDerivative(params map[string]interface{}, inputs []shader.Result) shader.Result {
	// TODO
	return shader.Result{}
}

func (tmap *TextureMapper) ViewDependent() bool {
	// XXX: Shouldn't this occasionally be true?
	return false
}

func (tmap *TextureMapper) Dependencies() []shader.Node { return []shader.Node{} }

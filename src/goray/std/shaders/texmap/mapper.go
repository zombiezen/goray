//
//	goray/std/shaders/texmap/mapper.go
//	goray
//
//	Created by Ross Light on 2011-04-02.
//

// The texmap package provides a shader node that performs texture mapping with various options.
package texmap

import (
	"os"

	"goray/core/matrix"
	"goray/core/render"
	"goray/core/shader"
	"goray/core/surface"
	"goray/core/vector"

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
}

var _ shader.Node = &TextureMapper{}

// New creates a new texture mapper with the given parameters.
func New(tex Texture, coord Coordinates, scalar bool) (tmap *TextureMapper) {
	tmap = new(TextureMapper)
	tmap.Init(tex, coord, scalar)
	return
}

// Init initializes the mapper with default values.  You do not *need* to call this method to use a texture mapper, but it does provide reasonable defaults.
func (tmap *TextureMapper) Init(tex Texture, coord Coordinates, scalar bool) {
	tmap.Texture = tex
	tmap.Coordinates = coord
	tmap.Projector = FlatMap
	tmap.MapX, tmap.MapY, tmap.MapZ = vector.X, vector.Y, vector.Z
	tmap.Transform = matrix.Identity
	tmap.Scale = vector.Vector3D{1.0, 1.0, 1.0}
	tmap.Offset = vector.Vector3D{}
	tmap.Scalar = scalar
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
	sp := params["SurfacePoint"].(surface.Point)
	state := params["RenderState"].(*render.State)
	p, n := sp.Position, sp.GeometricNormal
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

func (tmap *TextureMapper) EvalDerivative(inputs []shader.Result, params shader.Params) shader.Result {
	// TODO
	return shader.Result{}
}

func (tmap *TextureMapper) ViewDependent() bool {
	// XXX: Shouldn't this occasionally be true?
	return false
}

func (tmap *TextureMapper) Dependencies() []shader.Node { return []shader.Node{} }

func Construct(m yamldata.Map) (data interface{}, err os.Error) {
	tex, ok := m["texture"].(Texture)
	if !ok {
		err = os.NewError("Texture mapper must be given a texture")
		return
	}
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
	scalar, _ := yamldata.AsBool(m["scalar"])
	return New(tex, coord, scalar), nil
}

//
//	goray/std/materials/shinydiffuse.go
//	goray
//
//	Created by Ross Light on 2010-07-14.
//

package shinydiffuse

import (
	"os"
	"goray/core/color"
	"goray/core/material"
	"goray/core/render"
	"goray/core/shader"
	"goray/core/surface"
	"goray/core/vector"
	yamldata "goyaml.googlecode.com/hg/data"
)

type ShinyDiffuse struct {
	Color, SpecReflCol                color.Color
	Diffuse, SpecRefl, Transp, Transl float64
	TransmitFilter float64

	DiffuseShad, SpecReflShad, TranspShad, TranslShad, MirColShad shader.Node

	isDiffuse, isReflective, isTransp, isTransl bool

	fresnelEffect bool
	bsdfFlags     material.BSDF
	nBsdf         int
}

func (sd *ShinyDiffuse) Init() {
	sd.nBsdf = 0
	acc := 1.0
	if sd.SpecRefl > 0 || sd.SpecReflShad != nil {
		sd.isReflective = true
		// TODO: viNodes?
		sd.bsdfFlags |= material.BSDFSpecular | material.BSDFReflect
		sd.nBsdf++
	}
	// TODO: Transparency
	// TODO: Translucency
	if sd.Diffuse * acc > 0 {
		sd.isDiffuse = true
		// TODO: viNodes?
		sd.bsdfFlags |= material.BSDFDiffuse | material.BSDFReflect
		sd.nBsdf++
	}
}

type sdData struct {
	Components [4]float64
	Diffuse    shader.Result
	Transp     shader.Result
	Transl     shader.Result
	SpecRefl   shader.Result
	MirCol     shader.Result
}

func makeSdData(sd *ShinyDiffuse, use [4]bool) (data sdData) {
	params := make(map[string]interface{})
	results := shader.Eval(params, []shader.Node{sd.DiffuseShad, sd.TranspShad, sd.TranslShad, sd.SpecReflShad, sd.MirColShad})
	data.Diffuse = results[0]
	data.Transp = results[1]
	data.Transl = results[2]
	data.SpecRefl = results[3]
	data.MirCol = results[4]
	if sd.isReflective {
		if use[0] {
			data.Components[0] = data.SpecRefl.Scalar()
		} else {
			data.Components[0] = sd.SpecRefl
		}
	}
	if sd.isTransp {
		if use[1] {
			data.Components[1] = data.Transp.Scalar()
		} else {
			data.Components[1] = sd.Transp
		}
	}
	if sd.isTransl {
		if use[2] {
			data.Components[2] = data.Transl.Scalar()
		} else {
			data.Components[2] = sd.Transl
		}
	}
	if sd.isDiffuse {
		data.Components[3] = sd.Diffuse
	}
	return
}

func (sd *ShinyDiffuse) InitBSDF(state *render.State, sp surface.Point) material.BSDF {
	state.MaterialData = makeSdData(sd, [4]bool{false, false, false, false}) // TODO: viNodes...
	return sd.bsdfFlags
}

func (sd *ShinyDiffuse) GetFlags() material.BSDF {
	return sd.bsdfFlags
}

func (sd *ShinyDiffuse) getFresnel(wo, n vector.Vector3D) float64 {
	// TODO
	return 1.0
}

// calculate the absolute value of scattering components from the "normalized"
// fractions which are between 0 (no scattering) and 1 (scatter all remaining light)
// Kr is an optional reflection multiplier (e.g. from Fresnel)
func (sd *ShinyDiffuse) accumulate(components [4]float64, kr float64) (accum [4]float64) {
	accum[0] = components[0] * kr
	acc := 1 - accum[0]
	accum[1] = components[1] * acc
	acc *= 1 - components[1]
	accum[2] = components[2] * acc
	acc *= 1 - components[2]
	accum[3] = components[3] * acc
	return
}

func (sd *ShinyDiffuse) Eval(state *render.State, sp surface.Point, wo, wl vector.Vector3D, types material.BSDF) (col color.Color) {
	cosNgWo := vector.Dot(sp.GeometricNormal, wo)
	cosNgWl := vector.Dot(sp.GeometricNormal, wl)
	col = color.Black

	n := sp.Normal
	if cosNgWo < 0 {
		n = vector.ScalarMul(n, -1)
	}

	if types&sd.bsdfFlags&material.BSDFDiffuse == 0 {
		return
	}

	data := state.MaterialData.(sdData)

	kr := sd.getFresnel(wo, n)
	mt := (1 - kr*data.Components[0]) * (1 - data.Components[1])

	if cosNgWo*cosNgWl < 0 {
		// Transmit -- light comes from opposite side of surface
		if sd.isTransl {
			if sd.DiffuseShad != nil {
				col = data.Diffuse.Color()
			} else {
				col = sd.Color
			}
			col = color.ScalarMul(col, data.Components[2]*mt)
		}
		return
	}

	if vector.Dot(n, wl) < 0 {
		return
	}
	md := mt * (1 - data.Components[2]) * data.Components[3]
	// TODO: Oren-Nayer
	if sd.DiffuseShad != nil {
		col = data.Diffuse.Color()
	} else {
		col = sd.Color
	}
	col = color.ScalarMul(col, md)
	return
}

func (sd *ShinyDiffuse) Sample(state *render.State, sp surface.Point, wo vector.Vector3D, s *material.Sample) (col color.Color, wi vector.Vector3D) {
	// TODO
	col = color.Black
	return
}

func (sd *ShinyDiffuse) Pdf(state *render.State, sp surface.Point, wo, wi vector.Vector3D, bsdfs material.BSDF) float64 {
	// TODO
	return 0.0
}

func (sd *ShinyDiffuse) GetSpecular(state *render.State, sp surface.Point, wo vector.Vector3D) (reflect, refract bool, dir [2]vector.Vector3D, col [2]color.Color) {
	data := state.MaterialData.(sdData)
	backface := vector.Dot(sp.GeometricNormal, wo) < 0
	n, ng := sp.Normal, sp.GeometricNormal
	if backface {
		n = vector.ScalarMul(n, -1)
		ng = vector.ScalarMul(ng, -1)
	}
	kr := sd.getFresnel(wo, n)
	refract = sd.isTransp
	if sd.isTransp {
		dir[1] = vector.ScalarMul(wo, -1)
		if sd.DiffuseShad != nil {
			col[1] = data.Diffuse.Color()
		} else {
			col[1] = sd.Color
		}
		col[1] = color.Add(color.ScalarMul(col[1], sd.TransmitFilter), color.Gray(1 - sd.TransmitFilter))
		col[1] = color.ScalarMul(col[1], (1 - data.Components[0] * kr) * data.Components[1])
	}
	reflect = sd.isReflective
	if sd.isReflective {
		dir[0] = vector.Sub(vector.ScalarMul(n, 2.0 * vector.Dot(wo, n)), wo)
		cosWiNg := vector.Dot(dir[0], ng)
		if cosWiNg < 0.01 {
			dir[0] = vector.Add(dir[0], vector.ScalarMul(ng, (0.01 - cosWiNg)))
			dir[0] = dir[0].Normalize()
		}
		if sd.MirColShad != nil {
			col[0] = data.MirCol.Color()
		} else {
			col[0] = sd.SpecReflCol
		}
		col[0] = color.ScalarMul(col[0], data.Components[0] * kr)
	}
	return
}

func (sd *ShinyDiffuse) GetReflectivity(state *render.State, sp surface.Point, flags material.BSDF) color.Color {
	// TODO
	return color.Black
}

func (sd *ShinyDiffuse) GetAlpha(state *render.State, sp surface.Point, wo vector.Vector3D) float64 {
	if sd.isTransp {
		//data := state.MaterialData.(sdData)
	}
	return 1
}

func (sd *ShinyDiffuse) ScatterPhoton(state *render.State, sp surface.Point, wi vector.Vector3D, s *material.PhotonSample) (wo vector.Vector3D, scattered bool) {
	// TODO
	return
}

func Construct(m yamldata.Map) (data interface{}, err os.Error) {
	col, ok := m["color"].(color.Color)
	if !ok {
		err = os.NewError("Color must be an RGB")
		return
	}
	srcol, ok := m["mirrorColor"].(color.Color)
	if !ok {
		err = os.NewError("Mirror color must be an RGB")
		return
	}
	diffuse, ok := m["diffuseReflect"].(float64)
	if !ok {
		err = os.NewError("Diffuse reflection must be a float")
		return
	}
	specRefl, ok := m["specularReflect"].(float64)
	if !ok {
		// TODO: Better error checking
		specRefl = 0
	}

	mat := &ShinyDiffuse{
		Color: col,
		SpecReflCol: srcol,
		Diffuse: diffuse,
		SpecRefl: specRefl,
	}
	mat.Init()
	return mat, nil
}

//
//	goray/std/materials/shinydiffuse.go
//	goray
//
//	Created by Ross Light on 2010-07-14.
//

package shinydiffuse

import (
	"math"
	"os"
	"goray/core/color"
	"goray/core/material"
	"goray/core/render"
	"goray/core/shader"
	"goray/core/surface"
	"goray/core/vector"
	"goray/std/materials/common"
	yamldata "goyaml.googlecode.com/hg/data"
)

type ShinyDiffuse struct {
	Color, SpecReflCol                color.Color
	Diffuse, SpecRefl, Transp, Transl float64
	TransmitFilter                    float64
	IOR                               float64

	DiffuseShad, SpecReflShad, TranspShad, TranslShad, MirColShad shader.Node

	isDiffuse, isReflective, isTransp, isTransl bool

	fresnelEffect bool
	bsdfFlags     material.BSDF
}

func (sd *ShinyDiffuse) Init() {
	acc := 1.0
	if sd.SpecRefl > 0 || sd.SpecReflShad != nil {
		sd.isReflective = true
		// TODO: acc
		// TODO: viNodes?
		sd.bsdfFlags |= material.BSDFSpecular | material.BSDFReflect
	}
	// TODO: Transparency
	// TODO: Translucency
	if sd.Diffuse*acc > 0 {
		sd.isDiffuse = true
		// TODO: acc
		// TODO: viNodes?
		sd.bsdfFlags |= material.BSDFDiffuse | material.BSDFReflect
	}
}

type sdData struct {
	Diffuse, SpecRefl, Transp, Transl float64
	DiffuseColor, MirrorColor         color.Color
}

func makeSdData(sd *ShinyDiffuse, use [4]bool) (data sdData) {
	params := make(map[string]interface{})
	results := shader.Eval(params, []shader.Node{sd.DiffuseShad, sd.TranspShad, sd.TranslShad, sd.SpecReflShad, sd.MirColShad})
	if sd.isReflective {
		if use[0] {
			data.SpecRefl = results[3].Scalar()
		} else {
			data.SpecRefl = sd.SpecRefl
		}
	}
	if sd.isTransp {
		if use[1] {
			data.Transp = results[1].Scalar()
		} else {
			data.Transp = sd.Transp
		}
	}
	if sd.isTransl {
		if use[2] {
			data.Transl = results[2].Scalar()
		} else {
			data.Transl = sd.Transl
		}
	}
	if sd.isDiffuse {
		data.Diffuse = sd.Diffuse
	}
	if sd.DiffuseShad != nil {
		data.DiffuseColor = results[0].Color()
	} else {
		data.DiffuseColor = sd.Color
	}
	if sd.MirColShad != nil {
		data.MirrorColor = results[4].Color()
	} else {
		data.MirrorColor = sd.SpecReflCol
	}
	return
}

// calculate the absolute value of scattering components from the "normalized"
// fractions which are between 0 (no scattering) and 1 (scatter all remaining light)
// Kr is an optional reflection multiplier (e.g. from Fresnel)
func (data sdData) accumulate(kr float64) (newData sdData) {
	newData.DiffuseColor, newData.MirrorColor = data.DiffuseColor, data.MirrorColor

	newData.SpecRefl = data.SpecRefl * kr
	acc := 1 - newData.SpecRefl
	newData.Transp = data.Transp * acc
	acc *= 1 - data.Transp
	newData.Transp = data.Transl * acc
	acc *= 1 - data.Transl
	newData.Transp = data.Diffuse * acc
	return
}

func (sd *ShinyDiffuse) InitBSDF(state *render.State, sp surface.Point) material.BSDF {
	state.MaterialData = makeSdData(sd, [4]bool{false, false, false, false}) // TODO: viNodes...
	return sd.bsdfFlags
}

func (sd *ShinyDiffuse) GetFlags() material.BSDF {
	return sd.bsdfFlags
}

func (sd *ShinyDiffuse) getFresnel(wo, n vector.Vector3D) (kr float64) {
	if !sd.fresnelEffect {
		return 1.0
	}
	kr, _ = common.Fresnel(wo, n, sd.IOR)
	return
}

func (sd *ShinyDiffuse) Eval(state *render.State, sp surface.Point, wo, wl vector.Vector3D, types material.BSDF) (col color.Color) {
	cosNgWo := vector.Dot(sp.GeometricNormal, wo)
	cosNgWl := vector.Dot(sp.GeometricNormal, wl)
	col = color.Black

	n := sp.Normal
	if cosNgWo < 0 {
		n = n.Negate()
	}

	if types&sd.bsdfFlags&material.BSDFDiffuse == 0 {
		return
	}

	data := state.MaterialData.(sdData)

	kr := sd.getFresnel(wo, n)
	mt := (1 - kr*data.SpecRefl) * (1 - data.Transp)

	if cosNgWo*cosNgWl < 0 {
		// Transmit -- light comes from opposite side of surface
		if sd.isTransl {
			col = color.ScalarMul(data.DiffuseColor, data.Transl*mt)
		}
		return
	}

	if vector.Dot(n, wl) < 0 {
		return
	}
	md := mt * (1 - data.Transl) * data.Diffuse
	// TODO: Oren-Nayer
	col = color.ScalarMul(data.DiffuseColor, md)
	return
}

func (sd *ShinyDiffuse) Sample(state *render.State, sp surface.Point, wo vector.Vector3D, s *material.Sample) (col color.Color, wi vector.Vector3D) {
	// TODO
	col = color.Black
	return
}

func (sd *ShinyDiffuse) Pdf(state *render.State, sp surface.Point, wo, wi vector.Vector3D, bsdfs material.BSDF) (pdf float64) {
	if bsdfs&material.BSDFDiffuse == 0 {
		return
	}

	data := state.MaterialData.(sdData)
	cosNgWo := vector.Dot(sp.GeometricNormal, wo)
	cosNgWi := vector.Dot(sp.GeometricNormal, wi)
	n := sp.Normal
	if cosNgWo < 0 {
		n = n.Negate()
	}
	kr := sd.getFresnel(wo, n)
	accumC := data.accumulate(kr)
	possibleBsdfs := map[material.BSDF]float64{
		material.BSDFSpecular | material.BSDFReflect: accumC.SpecRefl,
		material.BSDFTransmit | material.BSDFFilter:  accumC.Transp,
		material.BSDFDiffuse | material.BSDFTransmit: accumC.Transl,
		material.BSDFDiffuse | material.BSDFReflect:  accumC.Diffuse,
	}

	sum := 0.0
	matched := false
	for p, width := range possibleBsdfs {
		if bsdfs&p == p {
			sum += width
			switch p {
			case material.BSDFDiffuse | material.BSDFTransmit: // translucency
				if cosNgWo*cosNgWi < 0 {
					pdf += math.Fabs(vector.Dot(wi, n)) * width
				}
			case material.BSDFDiffuse | material.BSDFReflect: // translucency
				if cosNgWo*cosNgWi > 0 {
					pdf += math.Fabs(vector.Dot(wi, n)) * width
				}
			}
			matched = true
		}
	}
	if !matched || sum < 0.00001 {
		return 0.0
	}
	return pdf / sum
}

func (sd *ShinyDiffuse) GetSpecular(state *render.State, sp surface.Point, wo vector.Vector3D) (reflect, refract bool, dir [2]vector.Vector3D, col [2]color.Color) {
	data := state.MaterialData.(sdData)
	backface := vector.Dot(sp.GeometricNormal, wo) < 0
	n, ng := sp.Normal, sp.GeometricNormal
	if backface {
		n, ng = n.Negate(), ng.Negate()
	}
	kr := sd.getFresnel(wo, n)
	refract = sd.isTransp
	if sd.isTransp {
		dir[1] = wo.Negate()
		col[1] = color.Add(color.ScalarMul(data.DiffuseColor, sd.TransmitFilter), color.Gray(1-sd.TransmitFilter))
		col[1] = color.ScalarMul(col[1], (1-data.SpecRefl*kr)*data.Transp)
	}
	reflect = sd.isReflective
	if sd.isReflective {
		dir[0] = vector.Sub(vector.ScalarMul(n, 2.0*vector.Dot(wo, n)), wo)
		cosWiNg := vector.Dot(dir[0], ng)
		if cosWiNg < 0.01 {
			dir[0] = vector.Add(dir[0], vector.ScalarMul(ng, (0.01-cosWiNg)))
			dir[0] = dir[0].Normalize()
		}
		col[0] = color.ScalarMul(data.MirrorColor, data.SpecRefl*kr)
	}
	return
}

func (sd *ShinyDiffuse) GetReflectivity(state *render.State, sp surface.Point, flags material.BSDF) color.Color {
	return common.GetReflectivity(sd, state, sp, flags)
}

func (sd *ShinyDiffuse) GetAlpha(state *render.State, sp surface.Point, wo vector.Vector3D) float64 {
	if sd.isTransp {
		//data := state.MaterialData.(sdData)
	}
	return 1
}

func (sd *ShinyDiffuse) ScatterPhoton(state *render.State, sp surface.Point, wi vector.Vector3D, s *material.PhotonSample) (wo vector.Vector3D, scattered bool) {
	return common.ScatterPhoton(sd, state, sp, wi, s)
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
		Color:       col,
		SpecReflCol: srcol,
		Diffuse:     diffuse,
		SpecRefl:    specRefl,
	}
	mat.Init()
	return mat, nil
}

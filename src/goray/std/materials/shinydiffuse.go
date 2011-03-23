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
	"goray/sampleutil"
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
	Color            color.Color
	Diffuse          float64
	DiffuseColorShad shader.Node

	SpecReflColor   color.Color
	SpecRefl        float64
	SpecReflShad    shader.Node
	MirrorColorShad shader.Node

	Transp, Transl         float64
	TranspShad, TranslShad shader.Node

	EmitColor color.Color
	EmitValue float64

	TransmitFilter, IOR float64

	isDiffuse, isReflective, isTransp, isTransl bool

	fresnelEffect bool
	bsdfFlags     material.BSDF
}

func (sd *ShinyDiffuse) Init() {
	const threshold = 1e-5

	sd.EmitColor = color.ScalarMul(sd.EmitColor, sd.EmitValue)
	acc := 1.0
	if sd.SpecRefl > threshold || sd.SpecReflShad != nil {
		sd.isReflective = true
		// TODO: viNodes?
		if sd.SpecReflShad == nil && !sd.fresnelEffect {
			acc = 1.0 - sd.SpecRefl
		}
		sd.bsdfFlags |= material.BSDFSpecular | material.BSDFReflect
	}
	if sd.Transp*acc > threshold || sd.TranspShad != nil {
		sd.isTransp = true
		// TODO: viNodes?
		if sd.TranspShad == nil && !sd.fresnelEffect {
			acc = 1.0 - sd.Transp
		}
		sd.bsdfFlags |= material.BSDFTransmit | material.BSDFFilter
	}
	if sd.Transl*acc > threshold || sd.TranslShad != nil {
		sd.isTransl = true
		// TODO: viNodes?
		if sd.TranslShad == nil && !sd.fresnelEffect {
			acc = 1.0 - sd.Transl
		}
		sd.bsdfFlags |= material.BSDFDiffuse | material.BSDFTransmit
	}
	if sd.Diffuse*acc > threshold {
		sd.isDiffuse = true
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
	results := shader.Eval(params, []shader.Node{sd.DiffuseColorShad, sd.TranspShad, sd.TranslShad, sd.SpecReflShad, sd.MirrorColorShad})
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
	if sd.DiffuseColorShad != nil {
		data.DiffuseColor = results[0].Color()
	} else {
		data.DiffuseColor = sd.Color
	}
	if sd.MirrorColorShad != nil {
		data.MirrorColor = results[4].Color()
	} else {
		data.MirrorColor = sd.SpecReflColor
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
	newData.Transl = data.Transl * acc
	acc *= 1 - data.Transl
	newData.Diffuse = data.Diffuse * acc
	return
}

func (sd *ShinyDiffuse) InitBSDF(state *render.State, sp surface.Point) material.BSDF {
	state.MaterialData = makeSdData(sd, [4]bool{false, false, false, false}) // TODO: viNodes...
	return sd.bsdfFlags
}

func (sd *ShinyDiffuse) MaterialFlags() material.BSDF {
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

type sdc struct {
	BSDF  material.BSDF
	Value float64
}

type sdComps []sdc

func (comps sdComps) Get(b material.BSDF) (v float64, ok bool) {
	for _, c := range comps {
		if c.BSDF == b {
			return c.Value, true
		}
	}
	return
}

func getComps(sd *ShinyDiffuse, data sdData) (comps sdComps) {
	comps = make(sdComps, 0, 4)
	if sd.isReflective {
		comps = append(comps, sdc{material.BSDFSpecular | material.BSDFReflect, data.SpecRefl})
	}
	if sd.isTransp {
		comps = append(comps, sdc{material.BSDFTransmit | material.BSDFFilter, data.Transp})
	}
	if sd.isTransl {
		comps = append(comps, sdc{material.BSDFDiffuse | material.BSDFTransmit, data.Transl})
	}
	if sd.isDiffuse {
		comps = append(comps, sdc{material.BSDFDiffuse | material.BSDFReflect, data.Diffuse})
	}
	return
}

func (sd *ShinyDiffuse) Sample(state *render.State, sp surface.Point, wo vector.Vector3D, s *material.Sample) (col color.Color, wi vector.Vector3D) {
	data := state.MaterialData.(sdData)
	cosNgWo := vector.Dot(sp.GeometricNormal, wo)
	cosNgWi := vector.Dot(sp.GeometricNormal, wi)
	n := sp.Normal
	if cosNgWo < 0 {
		n = n.Negate()
	}
	kr := sd.getFresnel(wo, n)
	accumC := data.accumulate(kr)

	// Pick component to sample
	sum := 0.0
	comps := getComps(sd, accumC)
	vals := make([]float64, len(comps))
	for i, c := range comps {
		sum += c.Value
		vals[i] = sum
	}
	if len(comps) == 0 || sum < 1e-6 {
		s.SampledFlags = material.BSDFNone
		s.Pdf = 0
		col = color.White
		return
	}
	pick := -1
	invSum := 1 / sum
	for i, _ := range comps {
		vals[i] *= invSum
		comps[i].Value *= invSum
		if s.S1 <= vals[i] && pick < 0 {
			pick = i
		}
	}
	if pick < 0 {
		pick = len(comps) - 1
	}
	s1 := s.S1 / comps[pick].Value
	if pick > 0 {
		s1 -= vals[pick-1] / comps[pick].Value
	}

	// Update sample information
	col = color.Black
	switch comps[pick].BSDF {
	case material.BSDFSpecular | material.BSDFReflect:
		wi = vector.Reflect(n, wo)
		s.Pdf = comps[pick].Value
		col = color.ScalarMul(accumC.MirrorColor, accumC.SpecRefl)
		if s.Reverse {
			s.PdfBack = s.Pdf
			s.ColorBack = color.ScalarDiv(col, math.Fabs(vector.Dot(sp.Normal, wo)))
		}
		col = color.ScalarDiv(col, math.Fabs(vector.Dot(sp.Normal, wi)))
	case material.BSDFTransmit | material.BSDFFilter:
		wi = wo.Negate()
		col = color.ScalarMul(color.Add(color.ScalarMul(accumC.DiffuseColor, sd.TransmitFilter), color.Gray(1-sd.TransmitFilter)), accumC.Transp)
		cosN := math.Fabs(vector.Dot(wi, n))
		if cosN < 1e-6 {
			s.Pdf = 0
		} else {
			col = color.ScalarDiv(col, cosN)
			s.Pdf = comps[pick].Value
		}
	case material.BSDFDiffuse | material.BSDFTransmit:
		wi = sampleutil.CosHemisphere(n.Negate(), sp.NormalU, sp.NormalV, s1, s.S2)
		if cosNgWo*cosNgWi < 0 {
			col = color.ScalarMul(accumC.DiffuseColor, accumC.Transl)
		}
		s.Pdf = math.Fabs(vector.Dot(wi, n)) * comps[pick].Value
	case material.BSDFDiffuse | material.BSDFReflect:
		fallthrough
	default:
		wi = sampleutil.CosHemisphere(n, sp.NormalU, sp.NormalV, s1, s.S2)
		if cosNgWo*cosNgWi > 0 {
			col = color.ScalarMul(accumC.DiffuseColor, accumC.Diffuse)
		}
		// TODO: if OrenNayer
		s.Pdf = math.Fabs(vector.Dot(wi, n)) * comps[pick].Value
	}
	s.SampledFlags = comps[pick].BSDF
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

	sum := 0.0
	comps := getComps(sd, accumC)
	for _, c := range comps {
		sum += c.Value
		switch c.BSDF {
		case material.BSDFDiffuse | material.BSDFTransmit:
			if cosNgWo*cosNgWi < 0 {
				pdf += math.Fabs(vector.Dot(wi, n)) * c.Value
			}
		case material.BSDFDiffuse | material.BSDFReflect:
			if cosNgWo*cosNgWi > 0 {
				pdf += math.Fabs(vector.Dot(wi, n)) * c.Value
			}
		}
	}
	if len(comps) == 0 || sum < 0.00001 {
		return 0.0
	}
	return pdf / sum
}

func (sd *ShinyDiffuse) Specular(state *render.State, sp surface.Point, wo vector.Vector3D) (reflect, refract bool, dir [2]vector.Vector3D, col [2]color.Color) {
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

func (sd *ShinyDiffuse) Reflectivity(state *render.State, sp surface.Point, flags material.BSDF) color.Color {
	return common.GetReflectivity(sd, state, sp, flags)
}

func (sd *ShinyDiffuse) Alpha(state *render.State, sp surface.Point, wo vector.Vector3D) float64 {
	if sd.isTransp {
		data := state.MaterialData.(sdData)
		n := sp.Normal
		if vector.Dot(sp.GeometricNormal, wo) < 0 {
			n = n.Negate()
		}
		kr := sd.getFresnel(wo, n)
		return 1 - (1-data.SpecRefl*kr)*data.Transp
	}
	return 1
}

func (sd *ShinyDiffuse) ScatterPhoton(state *render.State, sp surface.Point, wi vector.Vector3D, s *material.PhotonSample) (wo vector.Vector3D, scattered bool) {
	return common.ScatterPhoton(sd, state, sp, wi, s)
}

func (sd *ShinyDiffuse) Emit(state *render.State, sp surface.Point, wo vector.Vector3D) color.Color {
	if sd.DiffuseColorShad != nil {
		data := state.MaterialData.(sdData)
		return color.ScalarMul(data.DiffuseColor, sd.EmitValue)
	}
	return sd.EmitColor
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
	diffuse, ok := yamldata.AsFloat(m["diffuseReflect"])
	if !ok {
		err = os.NewError("Diffuse reflection must be a float")
		return
	}
	specRefl, ok := yamldata.AsFloat(m["specularReflect"])
	if !ok {
		// TODO: Better error checking
		specRefl = 0
	}
	transp, ok := yamldata.AsFloat(m["transparency"])
	if !ok {
		// TODO: Better error checking
		transp = 0
	}
	transl, ok := yamldata.AsFloat(m["translucency"])
	if !ok {
		// TODO: Better error checking
		transl = 0
	}
	transmit, ok := yamldata.AsFloat(m["transmit"])
	if !ok {
		// TODO: Better error checking
		transmit = 0
	}

	mat := &ShinyDiffuse{
		Color:          col,
		SpecReflColor:  srcol,
		EmitColor:      color.Black,
		Diffuse:        diffuse,
		SpecRefl:       specRefl,
		Transp:         transp,
		Transl:         transl,
		TransmitFilter: transmit,
	}
	mat.Init()
	return mat, nil
}

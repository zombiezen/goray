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

package materials

import (
	"errors"
	"math"

	"bitbucket.org/zombiezen/math3/vec64"
	"zombiezen.com/go/goray/internal/color"
	"zombiezen.com/go/goray/internal/goray"
	"zombiezen.com/go/goray/internal/sampleutil"
	"zombiezen.com/go/goray/internal/shader"
	yamldata "zombiezen.com/go/goray/internal/yaml/data"
	"zombiezen.com/go/goray/internal/yamlscene"
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
	bsdfFlags     goray.BSDF

	viewDependent bool
	useShaders    [4]bool
}

var (
	_ goray.Material     = &ShinyDiffuse{}
	_ goray.EmitMaterial = &ShinyDiffuse{}
)

// Init initializes sd's internal parameters. This must be called before using
// the material.
func (sd *ShinyDiffuse) Init() {
	const threshold = 1e-5

	sd.EmitColor = color.ScalarMul(sd.EmitColor, sd.EmitValue)
	acc := 1.0
	if sd.SpecRefl > threshold || sd.SpecReflShad != nil {
		sd.isReflective = true
		if sd.SpecReflShad != nil {
			if sd.SpecReflShad.ViewDependent() {
				sd.viewDependent = true
			}
			sd.useShaders[0] = true
		} else if !sd.fresnelEffect {
			acc = 1.0 - sd.SpecRefl
		}
		sd.bsdfFlags |= goray.BSDFSpecular | goray.BSDFReflect
	}
	if sd.Transp*acc > threshold || sd.TranspShad != nil {
		sd.isTransp = true
		if sd.TranspShad != nil {
			if sd.TranspShad.ViewDependent() {
				sd.viewDependent = true
			}
			sd.useShaders[1] = true
		} else {
			acc = 1.0 - sd.Transp
		}
		sd.bsdfFlags |= goray.BSDFTransmit | goray.BSDFFilter
	}
	if sd.Transl*acc > threshold || sd.TranslShad != nil {
		sd.isTransl = true
		if sd.TranslShad != nil {
			if sd.TranslShad.ViewDependent() {
				sd.viewDependent = true
			}
			sd.useShaders[2] = true
		} else {
			acc = 1.0 - sd.Transl
		}
		sd.bsdfFlags |= goray.BSDFDiffuse | goray.BSDFTransmit
	}
	if sd.Diffuse*acc > threshold {
		sd.isDiffuse = true
		if sd.DiffuseColorShad != nil {
			if sd.DiffuseColorShad.ViewDependent() {
				sd.viewDependent = true
			}
			sd.useShaders[3] = true
		}
		sd.bsdfFlags |= goray.BSDFDiffuse | goray.BSDFReflect
	}
}

type sdData struct {
	Diffuse, SpecRefl, Transp, Transl float64
	DiffuseColor, MirrorColor         color.Color
}

func makeSdData(sd *ShinyDiffuse, state *goray.RenderState, sp goray.SurfacePoint, use [4]bool, params shader.Params) (data sdData) {
	results := shader.Eval(
		[]shader.Node{
			sd.DiffuseColorShad,
			sd.TranspShad,
			sd.TranslShad,
			sd.SpecReflShad,
			sd.MirrorColorShad,
		},
		params,
	)
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

func (sd *ShinyDiffuse) InitBSDF(state *goray.RenderState, sp goray.SurfacePoint) goray.BSDF {
	params := shader.Params{
		"RenderState":  state,
		"SurfacePoint": sp,
	}
	if !sd.viewDependent {
		state.MaterialData = makeSdData(sd, state, sp, sd.useShaders, params)
	} else {
		// TODO: Allow view-dependent shaders
		state.MaterialData = makeSdData(sd, state, sp, [4]bool{}, params)
	}
	return sd.bsdfFlags
}

func (sd *ShinyDiffuse) MaterialFlags() goray.BSDF {
	return sd.bsdfFlags
}

func (sd *ShinyDiffuse) getFresnel(wo, n vec64.Vector) (kr float64) {
	if !sd.fresnelEffect {
		return 1.0
	}
	kr, _ = fresnel(wo, n, sd.IOR)
	return
}

func (sd *ShinyDiffuse) Eval(state *goray.RenderState, sp goray.SurfacePoint, wo, wl vec64.Vector, types goray.BSDF) (col color.Color) {
	cosNgWo := vec64.Dot(sp.GeometricNormal, wo)
	cosNgWl := vec64.Dot(sp.GeometricNormal, wl)
	col = color.Black

	n := sp.Normal
	if cosNgWo < 0 {
		n = n.Negate()
	}

	if types&sd.bsdfFlags&goray.BSDFDiffuse == 0 {
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

	if vec64.Dot(n, wl) < 0 {
		return
	}
	md := mt * (1 - data.Transl) * data.Diffuse
	// TODO: Oren-Nayer
	col = color.ScalarMul(data.DiffuseColor, md)
	return
}

type sdc struct {
	BSDF  goray.BSDF
	Value float64
}

type sdComps []sdc

func (comps sdComps) Get(b goray.BSDF) (v float64, ok bool) {
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
		comps = append(comps, sdc{goray.BSDFSpecular | goray.BSDFReflect, data.SpecRefl})
	}
	if sd.isTransp {
		comps = append(comps, sdc{goray.BSDFTransmit | goray.BSDFFilter, data.Transp})
	}
	if sd.isTransl {
		comps = append(comps, sdc{goray.BSDFDiffuse | goray.BSDFTransmit, data.Transl})
	}
	if sd.isDiffuse {
		comps = append(comps, sdc{goray.BSDFDiffuse | goray.BSDFReflect, data.Diffuse})
	}
	return
}

func (sd *ShinyDiffuse) Sample(state *goray.RenderState, sp goray.SurfacePoint, wo vec64.Vector, s *goray.MaterialSample) (col color.Color, wi vec64.Vector) {
	data := state.MaterialData.(sdData)
	cosNgWo := vec64.Dot(sp.GeometricNormal, wo)
	cosNgWi := vec64.Dot(sp.GeometricNormal, wi)
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
		s.SampledFlags = goray.BSDFNone
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
	case goray.BSDFSpecular | goray.BSDFReflect:
		wi = vec64.Reflect(n, wo)
		s.Pdf = comps[pick].Value
		col = color.ScalarMul(accumC.MirrorColor, accumC.SpecRefl)
		if s.Reverse {
			s.PdfBack = s.Pdf
			s.ColorBack = color.ScalarDiv(col, math.Abs(vec64.Dot(sp.Normal, wo)))
		}
		col = color.ScalarDiv(col, math.Abs(vec64.Dot(sp.Normal, wi)))
	case goray.BSDFTransmit | goray.BSDFFilter:
		wi = wo.Negate()
		col = color.ScalarMul(color.Add(color.ScalarMul(accumC.DiffuseColor, sd.TransmitFilter), color.Gray(1-sd.TransmitFilter)), accumC.Transp)
		cosN := math.Abs(vec64.Dot(wi, n))
		if cosN < 1e-6 {
			s.Pdf = 0
		} else {
			col = color.ScalarDiv(col, cosN)
			s.Pdf = comps[pick].Value
		}
	case goray.BSDFDiffuse | goray.BSDFTransmit:
		wi = sampleutil.CosHemisphere(n.Negate(), sp.NormalU, sp.NormalV, s1, s.S2)
		if cosNgWo*cosNgWi < 0 {
			col = color.ScalarMul(accumC.DiffuseColor, accumC.Transl)
		}
		s.Pdf = math.Abs(vec64.Dot(wi, n)) * comps[pick].Value
	case goray.BSDFDiffuse | goray.BSDFReflect:
		fallthrough
	default:
		wi = sampleutil.CosHemisphere(n, sp.NormalU, sp.NormalV, s1, s.S2)
		if cosNgWo*cosNgWi > 0 {
			col = color.ScalarMul(accumC.DiffuseColor, accumC.Diffuse)
		}
		// TODO: if OrenNayer
		s.Pdf = math.Abs(vec64.Dot(wi, n)) * comps[pick].Value
	}
	s.SampledFlags = comps[pick].BSDF
	return
}

func (sd *ShinyDiffuse) Pdf(state *goray.RenderState, sp goray.SurfacePoint, wo, wi vec64.Vector, bsdfs goray.BSDF) (pdf float64) {
	if bsdfs&goray.BSDFDiffuse == 0 {
		return
	}

	data := state.MaterialData.(sdData)
	cosNgWo := vec64.Dot(sp.GeometricNormal, wo)
	cosNgWi := vec64.Dot(sp.GeometricNormal, wi)
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
		case goray.BSDFDiffuse | goray.BSDFTransmit:
			if cosNgWo*cosNgWi < 0 {
				pdf += math.Abs(vec64.Dot(wi, n)) * c.Value
			}
		case goray.BSDFDiffuse | goray.BSDFReflect:
			if cosNgWo*cosNgWi > 0 {
				pdf += math.Abs(vec64.Dot(wi, n)) * c.Value
			}
		}
	}
	if len(comps) == 0 || sum < 0.00001 {
		return 0.0
	}
	return pdf / sum
}

func (sd *ShinyDiffuse) Specular(state *goray.RenderState, sp goray.SurfacePoint, wo vec64.Vector) (reflect, refract bool, dir [2]vec64.Vector, col [2]color.Color) {
	data := state.MaterialData.(sdData)
	backface := vec64.Dot(sp.GeometricNormal, wo) < 0
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
		dir[0] = vec64.Sub(n.Scale(2.0*vec64.Dot(wo, n)), wo)
		cosWiNg := vec64.Dot(dir[0], ng)
		if cosWiNg < 0.01 {
			dir[0] = vec64.Add(dir[0], ng.Scale(0.01-cosWiNg))
			dir[0] = dir[0].Normalize()
		}
		col[0] = color.ScalarMul(data.MirrorColor, data.SpecRefl*kr)
	}
	return
}

func (sd *ShinyDiffuse) Reflectivity(state *goray.RenderState, sp goray.SurfacePoint, flags goray.BSDF) color.Color {
	return getReflectivity(sd, state, sp, flags)
}

func (sd *ShinyDiffuse) Alpha(state *goray.RenderState, sp goray.SurfacePoint, wo vec64.Vector) float64 {
	if sd.isTransp {
		data := state.MaterialData.(sdData)
		n := sp.Normal
		if vec64.Dot(sp.GeometricNormal, wo) < 0 {
			n = n.Negate()
		}
		kr := sd.getFresnel(wo, n)
		return 1 - (1-data.SpecRefl*kr)*data.Transp
	}
	return 1
}

func (sd *ShinyDiffuse) ScatterPhoton(state *goray.RenderState, sp goray.SurfacePoint, wi vec64.Vector, s *goray.PhotonSample) (wo vec64.Vector, scattered bool) {
	return scatterPhoton(sd, state, sp, wi, s)
}

func (sd *ShinyDiffuse) Emit(state *goray.RenderState, sp goray.SurfacePoint, wo vec64.Vector) color.Color {
	if sd.DiffuseColorShad != nil {
		data := state.MaterialData.(sdData)
		return color.ScalarMul(data.DiffuseColor, sd.EmitValue)
	}
	return sd.EmitColor
}

func init() {
	yamlscene.Constructor[yamlscene.StdPrefix+"materials/shinydiffuse"] = yamlscene.MapConstruct(constructShinyDiffuse)
}

func constructShinyDiffuse(m yamldata.Map) (interface{}, error) {
	col, ok := m["color"].(color.Color)
	if !ok {
		return nil, errors.New("Color must be an RGB")
	}
	srcol, ok := m["mirrorColor"].(color.Color)
	if !ok {
		return nil, errors.New("Mirror color must be an RGB")
	}
	diffuse, ok := yamldata.AsFloat(m["diffuseReflect"])
	if !ok {
		return nil, errors.New("Diffuse reflection must be a float")
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

	diffuseColorShad, _ := m["diffuseColorShader"].(shader.Node)
	specReflShad, _ := m["specularReflectionShader"].(shader.Node)
	mirrorColorShad, _ := m["mirrorColorShader"].(shader.Node)
	transpShad, _ := m["transparencyShader"].(shader.Node)
	translShad, _ := m["translucencyShader"].(shader.Node)

	mat := &ShinyDiffuse{
		Color:            col,
		SpecReflColor:    srcol,
		EmitColor:        color.Black,
		Diffuse:          diffuse,
		SpecRefl:         specRefl,
		Transp:           transp,
		Transl:           transl,
		TransmitFilter:   transmit,
		DiffuseColorShad: diffuseColorShad,
		SpecReflShad:     specReflShad,
		MirrorColorShad:  mirrorColorShad,
		TranspShad:       transpShad,
		TranslShad:       translShad,
	}
	mat.Init()
	return mat, nil
}

//
//  goray/std/lights/point.go
//  goray
//
//  Created by Ross Light on 2010-06-02.
//

package point

import (
    "math"
    "./fmath"
)

import (
    "./goray/color"
    "./goray/light"
    "./goray/vector"
)

type PointLight struct {
    position vector.Vector3D
    color color.Color
    intensity float
}

func New(pos vector.Vector3D, col color.Color, intensity float) light.Light {
    pl := PointLight{position: pos, color: color.ScalarMul(col, intensity)}
    pl.intensity = color.GetEnergy(pl.color)
    return &pl
}

func (l *PointLight) Init(scene interface{}) {}

func (l *PointLight) TotalEnergy() color.Color {
}

func (l *PointLight) EmitPhoton(s1, s2, s3, s4 float) (col color.Color, r ray.Ray, ipdf float) {
    r.From = l.position
    r.Dir = sampleSphere(s1, s2)
    ipdf = 4.0 * math.Pi
    col = l.color
    return
}

func (l *PointLight) EmitSample(wo vector.Vector3D) (color.Color, Sample) {
}

func (l *PointLight) IllumSample(sp surface.Point, wi *ray.Ray) (s light.Sample, bool ok) {
    _, ok = l.Illuminate(sp, wi)
    if ok {
        s.Flags = l.GetFlags()
        s.Color = l.color
        s.Pdf = vector.Sub(l.position, sp.Position).LengthSqr()
    }
    return
}

func (l *PointLight) Illuminate(sp surface.Point, wi *ray.Ray) (col color.Color, ok bool) {
    ldir := vector.Sub(l.position, sp.Position)
    distSqr := ldir.LengthSqr()
    dist := fmath.Sqrt(distSqr)
    if dist == 0.0 {
        ok = false
        return
    }
    
    ok = true
    idistSqr := 1.0 / distSqr
    ldir = vector.ScalarMul(ldir, 1.0 / dist)
    
    wi.TMax = dist
    wi.Dir = ldir
    
    col = color.ScalarMul(l.color, idistSqr)
    return true
}

func (l *PointLight) CanIntersect() bool {
}

func (l *PointLight) Intersect(r ray.Ray) (ok bool, dist float, col color.Color, ipdf float) {
}

func (l *PointLight) IllumPdf(sp, spLight surface.Point) float {
}

func (l *PointLight) EmitPdf(sp surface.Point, wo vector.Vector3D) (areaPdf, dirPdf, cosWo float) {
}

func (l *PointLight) NumSamples() int {
}

func (l *PointLight) CanIlluminate(pt vector.Vector3D) bool {
}

func (l *PointLight) GetFlags() uint { return light.TypeSingular }

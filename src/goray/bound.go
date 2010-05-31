//
//  goray/bound.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

package bound

import (
	"./fmath"
	"./goray/vector"
)

type Bound struct {
	a, g vector.Vector3D
}

func New(a, g vector.Vector3D) *Bound {
	return &Bound{a, g}
}

func Union(b1, b2 *Bound) *Bound {
	newBound := &Bound{}
	newBound.a.X = fmath.Min(b1.a.X, b2.a.X)
	newBound.a.Y = fmath.Min(b1.a.Y, b2.a.Y)
	newBound.a.Z = fmath.Min(b1.a.Z, b2.a.Z)
	newBound.g.X = fmath.Max(b1.g.X, b2.g.X)
	newBound.g.Y = fmath.Max(b1.g.Y, b2.g.Y)
	newBound.g.Z = fmath.Max(b1.g.Z, b2.g.Z)
	return newBound
}

func (b *Bound) Get() (a, g vector.Vector3D) { return b.a, b.g }
func (b *Bound) GetMin() vector.Vector3D     { return b.a }
func (b *Bound) GetMax() vector.Vector3D     { return b.g }
func (b *Bound) Set(a, g vector.Vector3D)   { b.a = a; b.g = g }

func (b *Bound) SetMinX(x float) { b.a.X = x }
func (b *Bound) SetMinY(y float) { b.a.Y = y }
func (b *Bound) SetMinZ(z float) { b.a.Z = z }
func (b *Bound) SetMaxX(x float) { b.g.X = x }
func (b *Bound) SetMaxY(y float) { b.g.Y = y }
func (b *Bound) SetMaxZ(z float) { b.g.Z = z }

// Cross checks whether a given ray crosses the bound.
// from specifies a point where the ray starts.
// ray specifies the direction the ray is in.
// dist is the maximum distance that this method will check.  Pass in fmath.Inf
// to remove the check.
func (b *Bound) Cross(from, ray vector.Vector3D, dist float) (crosses bool, enter, leave float) {
	a0, a1 := b.a, b.g
	p := vector.Sub(from, a0)
	lmin, lmax := -1.0, -1.0

	if ray.X != 0 {
		tmp1 := -p.X / ray.X
		tmp2 := ((a1.X - a0.X) - p.X) / ray.X
		if tmp1 > tmp2 {
			tmp1, tmp2 = tmp2, tmp1
		}
		lmin, lmax = tmp1, tmp2
		if lmax < 0 || lmin > dist {
			return
		}
	}
	if ray.Y != 0 {
		tmp1 := -p.Y / ray.Y
		tmp2 := ((a1.Y - a0.Y) - p.Y) / ray.Y
		if tmp1 > tmp2 {
			tmp1, tmp2 = tmp2, tmp1
		}
		if tmp1 > lmin {
			lmin = tmp1
		}
		if tmp2 < lmax || lmax < 0 {
			lmax = tmp2
			if lmax < 0 || lmin > dist {
				return
			}
		}
	}
	if ray.Z != 0 {
		tmp1 := -p.Z / ray.Z
		tmp2 := ((a1.Z - a0.Z) - p.Z) / ray.Z
		if tmp1 > tmp2 {
			tmp1, tmp2 = tmp2, tmp1
		}
		if tmp1 > lmin {
			lmin = tmp1
		}
		if tmp2 < lmax || lmax < 0 {
			lmax = tmp2
		}
		if lmin <= lmax && lmax >= 0 && lmin <= dist {
			enter = lmin
			leave = lmax
			crosses = true
			return
		}
	}

	return
}

func (b *Bound) GetVolume() float {
	return (b.g.Y - b.a.Y) * (b.g.X - b.a.X) * (b.g.Z - b.a.Z)
}

func (b *Bound) GetXLength() float { return b.g.X - b.a.X }
func (b *Bound) GetYLength() float { return b.g.Y - b.a.Y }
func (b *Bound) GetZLength() float { return b.g.Z - b.a.Z }

func (b *Bound) Include(p vector.Vector3D) {
	b.a.X = fmath.Min(b.a.X, p.X)
	b.a.Y = fmath.Min(b.a.Y, p.Y)
	b.a.Z = fmath.Min(b.a.Z, p.Z)
	b.g.X = fmath.Max(b.g.X, p.X)
	b.g.Y = fmath.Max(b.g.Y, p.Y)
	b.g.Z = fmath.Max(b.g.Z, p.Z)
}

func (b *Bound) Includes(p vector.Vector3D) bool {
	return (p.X >= b.a.X && p.X <= b.g.X &&
		p.Y >= b.a.Y && p.Y <= b.g.Y &&
		p.Z >= b.a.Z && p.Z <= b.g.Z)
}

func (b *Bound) GetCenter() vector.Vector3D {
	return vector.ScalarMul(vector.Add(b.g, b.a), 0.5)
}

func (b *Bound) GetCenterX() float { return (b.g.X + b.a.X) * 0.5 }
func (b *Bound) GetCenterY() float { return (b.g.Y + b.a.Y) * 0.5 }
func (b *Bound) GetCenterZ() float { return (b.g.Z + b.a.Z) * 0.5 }

func (b *Bound) Grow(d float) {
	b.a.X -= d
	b.a.Y -= d
	b.a.Z -= d
	b.g.X += d
	b.g.Y += d
	b.g.Z += d
}

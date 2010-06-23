//
//  goray/core/bound/bound.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

/*
   The bound package provides a bounding box type, along with various
   manipulation operaions.
*/
package bound

import (
	"fmt"
	"goray/fmath"
	"goray/core/vector"
)

/*
   Bound is a simple bounding box.
   It should only be passed around as a pointer.
*/
type Bound struct {
	a, g vector.Vector3D
}

/* New creates a new bounding box from the two points given. */
func New(min, max vector.Vector3D) *Bound {
	return &Bound{min, max}
}

func (b *Bound) String() string {
	return fmt.Sprintf("Bound{min: %v, max: %v}", b.a, b.g)
}

/* Union creates a new bounding box that contains the two bounds. */
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

/* Get returns the minimum and maximum points that define the box. */
func (b *Bound) Get() (a, g vector.Vector3D) { return b.a, b.g }
/* Set changes the minimum and maximum points that define the box. */
func (b *Bound) Set(a, g vector.Vector3D) { b.a = a; b.g = g }

func (b *Bound) GetMin() vector.Vector3D { return b.a }
func (b *Bound) GetMax() vector.Vector3D { return b.g }

func (b *Bound) GetMinX() float { return b.a.X }
func (b *Bound) GetMinY() float { return b.a.Y }
func (b *Bound) GetMinZ() float { return b.a.Z }
func (b *Bound) GetMaxX() float { return b.g.X }
func (b *Bound) GetMaxY() float { return b.g.Y }
func (b *Bound) GetMaxZ() float { return b.g.Z }

func (b *Bound) SetMinX(x float) { b.a.X = x }
func (b *Bound) SetMinY(y float) { b.a.Y = y }
func (b *Bound) SetMinZ(z float) { b.a.Z = z }
func (b *Bound) SetMaxX(x float) { b.g.X = x }
func (b *Bound) SetMaxY(y float) { b.g.Y = y }
func (b *Bound) SetMaxZ(z float) { b.g.Z = z }

/*
   Cross checks whether a given ray crosses the bound.
   from specifies a point where the ray starts.
   ray specifies the direction the ray is in.
   dist is the maximum distance that this method will check.  Pass in fmath.Inf to remove the check.
*/
func (b *Bound) Cross(from, ray vector.Vector3D, dist float) (enter, leave float, crosses bool) {
	a0, a1 := b.a, b.g
	p := vector.Sub(from, a0)
	lmin, lmax := -1.0, -1.0

	if !fmath.Eq(ray.X, 0) {
		tmp1 := -p.X / ray.X
		tmp2 := ((a1.X - a0.X) - p.X) / ray.X
		if tmp1 > tmp2 {
			tmp1, tmp2 = tmp2, tmp1
		}
		lmin, lmax = tmp1, tmp2
		if lmax < 0 || lmin > dist {
			crosses = false
			return
		}
	}
	if !fmath.Eq(ray.Y, 0) {
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
				crosses = false
				return
			}
		}
	}
	if !fmath.Eq(ray.Z, 0) {
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
	}

	if lmin <= lmax && lmax >= 0 && lmin <= dist {
		enter, leave = lmin, lmax
		crosses = true
		return
	}

	crosses = false
	return
}

/* GetVolume calculates the volume of the bounding box */
func (b *Bound) GetVolume() float {
	return (b.g.Y - b.a.Y) * (b.g.X - b.a.X) * (b.g.Z - b.a.Z)
}

func (b *Bound) GetXLength() float { return b.g.X - b.a.X }
func (b *Bound) GetYLength() float { return b.g.Y - b.a.Y }
func (b *Bound) GetZLength() float { return b.g.Z - b.a.Z }

func (b *Bound) GetLargestAxis() int {
	x, y, z := b.GetXLength(), b.GetYLength(), b.GetZLength()
	switch {
	case z > y && z > x:
		return 2
	case y > z && y > x:
		return 1
	}
	return 0
}

func (b *Bound) GetHalfSize() [3]float {
	return [3]float{b.GetXLength() * 0.5, b.GetYLength() * 0.5, b.GetZLength() * 0.5}
}

/* Include modifies the bounding box so that it contains the specified point */
func (b *Bound) Include(p vector.Vector3D) {
	b.a.X = fmath.Min(b.a.X, p.X)
	b.a.Y = fmath.Min(b.a.Y, p.Y)
	b.a.Z = fmath.Min(b.a.Z, p.Z)
	b.g.X = fmath.Max(b.g.X, p.X)
	b.g.Y = fmath.Max(b.g.Y, p.Y)
	b.g.Z = fmath.Max(b.g.Z, p.Z)
}

/* Includes returns whether a given point is in the bounding box */
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

/* Grow increases the size of the bounding box by d on all sides.  The center will remain the same. */
func (b *Bound) Grow(d float) {
	b.a.X -= d
	b.a.Y -= d
	b.a.Z -= d
	b.g.X += d
	b.g.Y += d
	b.g.Z += d
}

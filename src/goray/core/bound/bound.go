//
//	goray/core/bound/bound.go
//	goray
//
//	Created by Ross Light on 2010-05-23.
//

/*
	The bound package provides a bounding box type, along with various
	manipulation operaions.
*/
package bound

import (
	"fmt"
	"math"
	"goray/core/vector"
)

// Bound is a simple bounding box.
// It should only be passed around as a pointer.
type Bound struct {
	a, g vector.Vector3D
}

// New creates a new bounding box from the two points given.
func New(min, max vector.Vector3D) *Bound {
	return &Bound{min, max}
}

func (b *Bound) String() string {
	return fmt.Sprintf("Bound{min: %v, max: %v}", b.a, b.g)
}

func (b *Bound) GoString() string {
	return fmt.Sprintf("Bound{a: %#v, g: %#v}", b.a, b.g)
}

// Union creates a new bounding box that contains the two bounds.
func Union(b1, b2 *Bound) *Bound {
	newBound := &Bound{}
	for axis := vector.X; axis <= vector.Z; axis++ {
		newBound.a[axis] = math.Fmin(b1.a[axis], b2.a[axis])
		newBound.g[axis] = math.Fmax(b1.g[axis], b2.g[axis])
	}
	return newBound
}

// Get returns the minimum and maximum points that define the box.
func (b *Bound) Get() (a, g vector.Vector3D) { return b.a, b.g }
// Set changes the minimum and maximum points that define the box.
func (b *Bound) Set(a, g vector.Vector3D) { b.a = a; b.g = g }

func (b *Bound) Min() vector.Vector3D { return b.a }
func (b *Bound) Max() vector.Vector3D { return b.g }

func (b *Bound) MinX() float64 { return b.a[vector.X] }
func (b *Bound) MinY() float64 { return b.a[vector.Y] }
func (b *Bound) MinZ() float64 { return b.a[vector.Z] }
func (b *Bound) MaxX() float64 { return b.g[vector.X] }
func (b *Bound) MaxY() float64 { return b.g[vector.Y] }
func (b *Bound) MaxZ() float64 { return b.g[vector.Z] }

func (b *Bound) SetMinX(x float64) { b.a[vector.X] = x }
func (b *Bound) SetMinY(y float64) { b.a[vector.Y] = y }
func (b *Bound) SetMinZ(z float64) { b.a[vector.Z] = z }
func (b *Bound) SetMaxX(x float64) { b.g[vector.X] = x }
func (b *Bound) SetMaxY(y float64) { b.g[vector.Y] = y }
func (b *Bound) SetMaxZ(z float64) { b.g[vector.Z] = z }

// Cross checks whether a given ray crosses the bound.
// from specifies a point where the ray starts.
// ray specifies the direction the ray is in.
// dist is the maximum distance that this method will check.  Pass in math.Inf(1) to remove the check.
func (b *Bound) Cross(from, ray vector.Vector3D, dist float64) (lmin, lmax float64, crosses bool) {
	p := vector.Sub(from, b.a)
	lmin, lmax = math.Inf(-1), math.Inf(1)
	crosses = true

	for axis := vector.X; axis <= vector.Z; axis++ {
		if ray[axis] != 0 {
			tmp1 := -p[axis] / ray[axis]
			tmp2 := (b.g[axis] - b.a[axis] - p[axis]) / ray[axis]
			if tmp1 > tmp2 {
				tmp1, tmp2 = tmp2, tmp1
			}
			if tmp1 > lmin {
				lmin = tmp1
			}
			if tmp2 < lmax {
				lmax = tmp2
			}
			if lmax < 0 || lmin > dist {
				return 0, 0, false
			}
		}
	}

	if lmin > lmax || lmax < 0 || lmin > dist {
		return 0, 0, false
	}
	return
}

// Volume calculates the volume of the bounding box
func (b *Bound) Volume() float64 {
	return (b.g[vector.Y] - b.a[vector.Y]) * (b.g[vector.X] - b.a[vector.X]) * (b.g[vector.Z] - b.a[vector.Z])
}

func (b *Bound) Size() [3]float64 { return [3]float64(vector.Sub(b.g, b.a)) }
func (b *Bound) LengthX() float64 { return b.g[vector.X] - b.a[vector.X] }
func (b *Bound) LengthY() float64 { return b.g[vector.Y] - b.a[vector.Y] }
func (b *Bound) LengthZ() float64 { return b.g[vector.Z] - b.a[vector.Z] }

func (b *Bound) LargestAxis() vector.Axis {
	s := b.Size()
	return vector.LargestAxis(s[vector.X], s[vector.Y], s[vector.Z])
}

func (b *Bound) HalfSize() [3]float64 {
	return [3]float64{b.LengthX() * 0.5, b.LengthY() * 0.5, b.LengthZ() * 0.5}
}

// Include modifies the bounding box so that it contains the specified point.
func (b *Bound) Include(p vector.Vector3D) {
	for axis := vector.X; axis <= vector.Z; axis++ {
		b.a[axis] = math.Fmin(b.a[axis], p[axis])
		b.g[axis] = math.Fmax(b.g[axis], p[axis])
	}
}

// Includes returns whether a given point is in the bounding box.
func (b *Bound) Includes(p vector.Vector3D) bool {
	for axis := vector.X; axis <= vector.Z; axis++ {
		if p[axis] < b.a[axis] || p[axis] > b.g[axis] {
			return false
		}
	}
	return true
}

func (b *Bound) Center() vector.Vector3D {
	return vector.ScalarMul(vector.Add(b.g, b.a), 0.5)
}

func (b *Bound) CenterX() float64 { return (b.g[vector.X] + b.a[vector.X]) * 0.5 }
func (b *Bound) CenterY() float64 { return (b.g[vector.Y] + b.a[vector.Y]) * 0.5 }
func (b *Bound) CenterZ() float64 { return (b.g[vector.Z] + b.a[vector.Z]) * 0.5 }

// Grow increases the size of the bounding box by d on all sides.  The center will remain the same.
func (b *Bound) Grow(d float64) {
	for axis := vector.X; axis <= vector.Z; axis++ {
		b.a[axis] -= d
		b.g[axis] += d
	}
}

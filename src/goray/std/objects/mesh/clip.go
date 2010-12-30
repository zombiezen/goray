//
//	goray/std/objects/mesh/clip.go
//	goray
//
//	Created by Ross Light on 2010-12-26.
//

package mesh

import (
	"math"
	"goray/core/bound"
	"goray/core/vector"
)

func calcBound(poly []vector.Vector3D) *bound.Bound {
	a, g := poly[0], poly[0]
	for i := 1; i < len(poly); i++ {
		for axis := vector.X; axis <= vector.Z; axis++ {
			a[axis] = math.Fmin(a[axis], poly[i][axis])
			g[axis] = math.Fmax(g[axis], poly[i][axis])
		}
	}
	return bound.New(a, g)
}

func triBoxClip(bMin, bMax [3]float64, poly []vector.Vector3D) ([]vector.Vector3D, *bound.Bound) {
	for axis := vector.X; axis <= vector.Z; axis++ { // for each axis
		// clip lower bound
		poly = triClip(axis, bMin[axis], poly, cmpMin)
		if len(poly) > 9 {
			// fatal error
			panic("clipped polygon is too complex")
		}
		if len(poly) == 0 {
			// entire polygon clipped
			return nil, nil
		}

		// clip upper bound
		poly = triClip(axis, bMax[axis], poly, cmpMax)
		if len(poly) > 10 {
			// fatal error
			panic("clipped polygon is too complex")
		}
		if len(poly) == 0 {
			// entire polygon clipped
			return nil, nil
		}
	}

	if len(poly) < 3 {
		panic("clipped polygon degenerated")
	}

	return poly, calcBound(poly)
}

func triPlaneClip(axis vector.Axis, pos float64, lower bool, poly []vector.Vector3D) ([]vector.Vector3D, *bound.Bound) {
	if lower {
		poly = triClip(axis, pos, poly, cmpMin)
	} else {
		poly = triClip(axis, pos, poly, cmpMax)
	}

	switch {
	case len(poly) == 0:
		return nil, nil
	case len(poly) < 3:
		panic("clipped polygon degenerated")
	case len(poly) > 10:
		panic("clipped polygon is too complex")
	}

	return poly, calcBound(poly)
}

// triClip is the internal clipping function. It's not very user-friendly; use triBoxClip or triPlaneClip.
func triClip(axis vector.Axis, bound float64, poly []vector.Vector3D, cmp func(a, b float64) bool) (cpoly []vector.Vector3D) {
	nextAxis, prevAxis := (axis+1)%3, (axis+2)%3

	cpoly = make([]vector.Vector3D, 0, 11)
	p1_inside := poly[0][axis] == bound || cmp(poly[0][axis], bound)

	for i := 0; i < len(poly)-1; i++ {
		p1, p2 := poly[i], poly[i+1]

		if p1_inside {
			if p2[axis] == bound || cmp(p2[axis], bound) {
				// both "inside"; copy p2 to new poly
				cpoly = append(cpoly, p2)
				p1_inside = true
			} else {
				// clip line, add intersection to new poly
				t := (bound - p1[axis]) / (p2[axis] - p1[axis])
				dv := vector.Vector3D{}
				dv[axis] = bound
				dv[nextAxis] = p2[nextAxis] + t*(p1[nextAxis]-p2[nextAxis])
				dv[prevAxis] = p2[prevAxis] + t*(p1[prevAxis]-p2[prevAxis])
				cpoly = append(cpoly, dv)
				p1_inside = false
			}
		} else {
			// p1 outside
			switch {
			case cmp(p2[axis], bound):
				// p2 inside, add s and p2
				t := (bound - p2[axis]) / (p1[axis] - p2[axis])
				dv := vector.Vector3D{}
				dv[axis] = bound
				dv[nextAxis] = p2[nextAxis] + t*(p1[nextAxis]-p2[nextAxis])
				dv[prevAxis] = p2[prevAxis] + t*(p1[prevAxis]-p2[prevAxis])
				cpoly = append(cpoly, dv, p2)
				p1_inside = true
			case p2[axis] == bound:
				// p2 and s are identical, only add p2
				cpoly = append(cpoly, p2)
				p1_inside = true
			default:
				// Both outside, do nothing
				p1_inside = false
			}
		}
	}

	if len(cpoly) > 0 {
		cpoly = append(cpoly, poly[0])
	}

	return
}

func cmpMin(a, b float64) bool { return a > b }
func cmpMax(a, b float64) bool { return a < b }

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

type dVector [3]float64

func vec2dvec(v vector.Vector3D) dVector {
	return dVector{float64(v.X), float64(v.Y), float64(v.Z)}
}

func dvec2vec(v dVector) vector.Vector3D {
	return vector.New(float(v[0]), float(v[1]), float(v[2]))
}

func triBoxClip(bMin, bMax [3]float64, poly []dVector) ([]dVector, *bound.Bound) {
	for axis := 0; axis < 3; axis++ { // for each axis
		// clip lower bound
		poly = triClip(axis, bMin[axis], poly, cmpMin)
		if len(poly) > 9 {
			// fatal error
			panic("clipped polygon is too complex")
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

	a, g := poly[0], poly[0]
	for i := 1; i < len(poly); i++ {
		a[0] = math.Fmin(a[0], poly[i][0])
		a[1] = math.Fmin(a[1], poly[i][1])
		a[2] = math.Fmin(a[2], poly[i][2])

		g[0] = math.Fmax(g[0], poly[i][0])
		g[1] = math.Fmax(g[1], poly[i][1])
		g[2] = math.Fmax(g[2], poly[i][2])
	}

	return poly, bound.New(dvec2vec(a), dvec2vec(g))
}

func triPlaneClip(axis int, pos float64, lower bool, poly []dVector) ([]dVector, *bound.Bound) {
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

	a, g := poly[0], poly[0]
	for i := 1; i < len(poly); i++ {
		a[0] = math.Fmin(a[0], poly[i][0])
		a[1] = math.Fmin(a[1], poly[i][1])
		a[2] = math.Fmin(a[2], poly[i][2])

		g[0] = math.Fmax(g[0], poly[i][0])
		g[1] = math.Fmax(g[1], poly[i][1])
		g[2] = math.Fmax(g[2], poly[i][2])
	}

	return poly, bound.New(dvec2vec(a), dvec2vec(g))
}

// triClip is the internal clipping function. It's not very user-friendly; use triBoxClip or triPlaneClip.
func triClip(axis int, bound float64, poly []dVector, cmp func(a, b float64) bool) (cpoly []dVector) {
	nextAxis, prevAxis := (axis+1)%3, (axis+2)%3

	cpoly = make([]dVector, 0, 11)
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
				dv := dVector{}
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
				dv := dVector{}
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

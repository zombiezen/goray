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

package kdtree

import (
	"bitbucket.org/zombiezen/goray/bound"
	"bitbucket.org/zombiezen/goray/vecutil"
	"math"
	"sort"
)

func split(indices []int, bd bound.Bound, state *buildState) (axis vecutil.Axis, pivot float64, cost float64) {
	const pigeonThreshold = 128
	if len(indices) > pigeonThreshold {
		return pigeonSplit(indices, bd, state)
	}
	return minimalSplit(indices, bd, state)
}

func simpleSplit(indices []int, bd bound.Bound, state *buildState) (axis vecutil.Axis, pivot float64, cost float64) {
	axis = bd.LargestAxis()
	data := make([]float64, 0, len(indices)*2)
	for _, i := range indices {
		min, max := state.ClippedDimension(i, axis)
		if min == max {
			data = append(data, min)
		} else {
			data = append(data, min, max)
		}
	}
	sort.Float64s(data)
	pivot = data[len(data)/2]
	return
}

func pigeonSplit(indices []int, bd bound.Bound, state *buildState) (bestAxis vecutil.Axis, bestPivot float64, bestCost float64) {
	const numBins = 1024
	type pigeonBin struct {
		n           int
		left, right int
		bleft, both int
		t           float64
	}

	var bins [numBins + 1]pigeonBin
	d := bd.Size()
	bestCost = math.Inf(1)
	totalSA := d[0]*d[1] + d[0]*d[2] + d[1]*d[2]
	invTotalSA := 0.0
	if totalSA != 0.0 {
		invTotalSA = 1.0 / totalSA
	}

	for axis := vecutil.X; axis <= vecutil.Z; axis++ {
		s := numBins / d[axis]
		min := bd.Min[axis]

		for _, i := range indices {
			tLow, tHigh := state.ClippedDimension(i, axis)
			bLeft, bRight := int((tLow-min)*s), int((tHigh-min)*s)
			if bLeft < 0 {
				bLeft = 0
			} else if bLeft > numBins {
				bLeft = numBins
			}
			if bRight < 0 {
				bRight = 0
			} else if bRight > numBins {
				bRight = numBins
			}

			if tLow == tHigh {
				if bins[bLeft].n == 0 || tLow >= bins[bLeft].t {
					bins[bLeft].t = tLow
					bins[bLeft].both++
				} else {
					bins[bLeft].left++
					bins[bLeft].right++
				}
				bins[bLeft].n += 2
			} else {
				if bins[bLeft].n == 0 || tLow > bins[bLeft].t {
					bins[bLeft].t = tLow
					bins[bLeft].left += bins[bLeft].both + bins[bLeft].bleft
					bins[bLeft].right += bins[bLeft].both
					bins[bLeft].both, bins[bLeft].bleft = 0, 0
					bins[bLeft].bleft++
				} else if tLow == bins[bLeft].t {
					bins[bLeft].bleft++
				} else {
					bins[bLeft].left++
				}

				bins[bLeft].n++
				bins[bRight].right++
				if bins[bRight].n == 0 || tHigh > bins[bRight].t {
					bins[bRight].t = tHigh
					bins[bRight].left += bins[bRight].both + bins[bRight].bleft
					bins[bRight].right += bins[bRight].both
					bins[bRight].both, bins[bRight].bleft = 0, 0
				}
				bins[bRight].n++
			}
		}

		capArea := d[axis.Next()] * d[axis.Prev()]
		capPerim := d[axis.Next()] + d[axis.Prev()]

		// Cumulate values and evaluate cost
		nBelow, nAbove := 0, len(indices)
		for _, b := range bins {
			if b.n != 0 {
				nBelow += b.left
				nAbove -= b.right
				// Cost:
				edget := b.t
				if edget > bd.Min[axis] && edget < bd.Max[axis] {
					cost := computeCost(axis, bd, capArea, capPerim, invTotalSA, nBelow, nAbove, edget)
					if cost < bestCost {
						bestAxis, bestPivot, bestCost = axis, edget, cost
					}
				}

				nBelow += b.both + b.bleft
				nAbove -= b.both
			}
		}

		if nBelow != len(indices) || nAbove != 0 {
			panic("Cost function mismatch")
		}

		// Reset all bins
		for i, _ := range bins {
			bins[i] = pigeonBin{}
		}
	}

	return
}

func computeCost(axis vecutil.Axis, bd bound.Bound, capArea, capPerim, invTotalSA float64, nBelow, nAbove int, edget float64) float64 {
	const emptyBonus = 0.33
	const costRatio = 0.35

	l1, l2 := edget-bd.Min[axis], bd.Max[axis]-edget
	belowSA, aboveSA := capArea+l1*capPerim, capArea+l2*capPerim
	rawCosts := belowSA*float64(nBelow) + aboveSA*float64(nAbove)

	d := bd.Size()[axis]

	var eb float64
	if nAbove == 0 {
		eb = (0.1 + l2/d) * emptyBonus * rawCosts
	} else if nBelow == 0 {
		eb = (0.1 + l1/d) * emptyBonus * rawCosts
	}

	return costRatio + invTotalSA*(rawCosts-eb)
}

type boundEdge struct {
	position float64
	boundEnd int
}

type boundEdgeArray []boundEdge

func (a boundEdgeArray) Len() int {
	return len(a)
}

func (a boundEdgeArray) Swap(i1, i2 int) {
	a[i1], a[i2] = a[i2], a[i1]
}

func (a boundEdgeArray) Less(i1, i2 int) bool {
	e, f := a[i1], a[i2]
	if e.position == f.position {
		return e.boundEnd > f.boundEnd
	}
	return e.position < f.position
}

const (
	lowerB = iota
	bothB
	upperB
)

func minimalSplit(indices []int, bd bound.Bound, state *buildState) (bestAxis vecutil.Axis, bestPivot float64, bestCost float64) {
	d := bd.Size()
	bestCost = math.Inf(1)
	totalSA := d[0]*d[1] + d[0]*d[2] + d[1]*d[2]
	invTotalSA := 0.0
	if totalSA != 0.0 {
		invTotalSA = 1.0 / totalSA
	}

	for axis := vecutil.X; axis <= vecutil.Z; axis++ {
		edges := make(boundEdgeArray, 0, len(indices)*2)
		for _, i := range indices {
			min, max := state.ClippedDimension(i, axis)
			if min == max {
				edges = append(edges, boundEdge{min, bothB})
			} else {
				edges = append(edges, boundEdge{min, lowerB}, boundEdge{max, upperB})
			}
		}
		sort.Sort(edges)

		capArea := d[axis.Next()] * d[axis.Prev()]
		capPerim := d[axis.Next()] + d[axis.Prev()]

		nBelow, nAbove := 0, len(indices)
		for _, e := range edges {
			if e.boundEnd == upperB {
				nAbove--
			}

			if e.position > bd.Min[axis] && e.position < bd.Max[axis] {
				cost := computeCost(axis, bd, capArea, capPerim, invTotalSA, nAbove, nBelow, e.position)
				if cost < bestCost {
					bestAxis, bestPivot, bestCost = axis, e.position, cost
				}
			}

			if e.boundEnd != upperB {
				nBelow++
				if e.boundEnd == bothB {
					nAbove--
				}
			}
		}

		if nBelow != len(indices) || nAbove != 0 {
			panic("Cost function mismatch")
		}
	}

	return
}

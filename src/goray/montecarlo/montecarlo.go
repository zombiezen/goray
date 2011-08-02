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

// Package montecarlo provides numerical algorithms helpful for performing
// Monte Carlo approximations.
package montecarlo

// Halton is a fast, incremental Halton sequence generator.
type Halton struct {
	base    uint
	invBase float64
	value   float64
}

// NewHalton creates a new Halton sequence generator.
func NewHalton(base uint) (h *Halton) {
	h = new(Halton)
	h.SetBase(base)
	return
}

// SetBase changes the Halton generator's base number.  This will reset the generator.
func (h *Halton) SetBase(base uint) {
	h.base, h.invBase = base, 1.0/float64(base)
	h.Reset()
}

// Reset starts the generator over without changing the base.
func (h *Halton) Reset() { h.value = 0.0 }

// SetStart changes the generator to start with the given index in the sequence.
func (h *Halton) SetStart(i uint) {
	h.value = 0
	for f, factor := h.invBase, h.invBase; i > 0; {
		h.value += float64(i%h.base) * factor
		i /= h.base
		factor *= f
	}
}

// Float64 returns the next number in the sequence with 64 bits of precision.
func (hal *Halton) Float64() float64 {
	r := 1 - hal.value - 1e-10
	if hal.invBase < r {
		hal.value += hal.invBase
	} else {
		hh, h := hal.invBase, hal.invBase*hal.invBase
		for h >= r {
			hh, h = h, h*hal.invBase
		}
		hal.value += hh + h - 1
	}
	return hal.value
}

// VanDerCorput returns the next number in the van der Corput sequence.
//
// This function can also be used to generate Sobol sequences and Larcher &
// Pillichshammer sequences.  The particular algorithm used is described in
// "Efficient Multidimensional Sampling" by Alexander Keller.
func VanDerCorput(bits, r uint32) float64 {
	const multRatio = 0.00000000023283064365386962890625

	bits = bits<<16 | bits>>16
	bits = bits&0x00ff00ff<<8 | bits&0xff00ff00>>8
	bits = bits&0x0f0f0f0f<<4 | bits&0xf0f0f0f0>>4
	bits = bits&0x33333333<<2 | bits&0xcccccccc>>2
	bits = bits&0x55555555<<1 | bits&0xaaaaaaaa>>1
	return float64(bits^r) * multRatio
}

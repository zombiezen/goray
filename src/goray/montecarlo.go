//
//	goray/montecarlo.go
//	goray
//
//	Created by Ross Light on 2010-06-16.
//

/*
	The montecarlo package provides numerical algorithms helpful for performing
	Monte Carlo approximations.
*/
package montecarlo

/* Halton is a fast, incremental Halton sequence generator. */
type Halton struct {
	base    uint
	invBase float64
	value   float64
}

/* NewHalton creates a new Halton sequence generator. */
func NewHalton(base uint) (h *Halton) {
	h = new(Halton)
	h.SetBase(base)
	return
}

/* SetBase changes the Halton generator's base number.  This will reset the generator. */
func (h *Halton) SetBase(base uint) {
	h.base, h.invBase = base, 1.0/float64(base)
	h.Reset()
}

/* Reset starts the generator over without changing the base. */
func (h *Halton) Reset() { h.value = 0.0 }

/* SetStart changes the generator to start with the given index in the sequence. */
func (h *Halton) SetStart(i uint) {
	h.value = 0
	for f, factor := h.invBase, h.invBase; i > 0; {
		h.value += float64(i%h.base) * factor
		i /= h.base
		factor *= f
	}
}

/* Float64 returns the next number in the sequence with 64 bits of precision. */
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

func (h *Halton) Float() float { return float(h.Float64()) }

/*
	VanDerCorput returns the next number in the van der Corput sequence.

	This function can also be used to generate Sobol sequences and Larcher &
	Pillichshammer sequences.  The particular algorithm used is described in
	"Efficient Multidimensional Sampling" by Alexander Keller.
*/
func VanDerCorput(bits, r uint32) float {
	const multRatio = 0.00000000023283064365386962890625

	bits = bits<<16 | bits>>16
	bits = bits&0x00ff00ff<<8 | bits&0xff00ff00>>8
	bits = bits&0x0f0f0f0f<<4 | bits&0xf0f0f0f0>>4
	bits = bits&0x33333333<<2 | bits&0xcccccccc>>2
	bits = bits&0x55555555<<1 | bits&0xaaaaaaaa>>1
	return float(float64(bits^r) * multRatio)
}

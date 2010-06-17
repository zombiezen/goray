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

/* SetStart changes the generator to start with a given number. */
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

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

package goray

import (
	"runtime"
	"sync"

	"bitbucket.org/zombiezen/math3/vec64"
	"zombiezen.com/go/goray/internal/color"
	"zombiezen.com/go/goray/internal/log"
)

// An Integrator renders rays.
type Integrator interface {
	Preprocess(scene *Scene)
	Integrate(scene *Scene, state *RenderState, r DifferentialRay) color.AlphaColor
}

// A SurfaceIntegrator renders rays by casting onto solid objects.
type SurfaceIntegrator interface {
	Integrator
	SurfaceIntegrator()
}

// A VolumeIntegrator renders rays by casting onto volumetric regions.
type VolumeIntegrator interface {
	Integrator
	VolumeIntegrator()
	Transmittance(scene *Scene, state *RenderState, r Ray) color.AlphaColor
}

// Render is an easy way of creating an image from a scene.
//
// Render will update the scene, create a new image, and then use one of the
// integration functions to write to the image.
func Render(s *Scene, i Integrator, log log.Logger) (img *Image) {
	s.Update()
	img = NewImage(s.Camera().ResolutionX(), s.Camera().ResolutionY())
	i.Preprocess(s)
	ch := BlockIntegrate(s, i, log)
	img.Acquire(ch)
	return
}

// RenderPixel creates a fragment for a position in the image.
func RenderPixel(s *Scene, i Integrator, x, y int) Fragment {
	cam := s.Camera()
	w, h := cam.ResolutionX(), cam.ResolutionY()

	// Set up state
	state := new(RenderState)
	state.Init()
	state.PixelNumber = y*w + x
	state.ScreenPos = vec64.Vector{2.0*float64(x)/float64(w) - 1.0, -2.0*float64(y)/float64(h) + 1.0, 0.0}
	state.Time = 0.0

	// Shoot ray
	r, _ := cam.ShootRay(float64(x), float64(y), 0, 0)

	// Set up differentials
	cRay := DifferentialRay{Ray: r}
	r, _ = cam.ShootRay(float64(x+1), float64(y), 0, 0)
	cRay.FromX = r.From
	cRay.DirX = r.Dir
	r, _ = cam.ShootRay(float64(x), float64(y+1), 0, 0)
	cRay.FromY = r.From
	cRay.DirY = r.Dir

	// Integrate
	color := i.Integrate(s, state, cRay)
	return Fragment{X: x, Y: y, Color: color}
}

const fragBufferSize = 100

// SimpleIntegrate integrates an image one pixel at a time.
func SimpleIntegrate(s *Scene, in Integrator, log log.Logger) <-chan Fragment {
	ch := make(chan Fragment, fragBufferSize)
	go func() {
		defer close(ch)
		w, h := s.Camera().ResolutionX(), s.Camera().ResolutionY()
		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				ch <- RenderPixel(s, in, x, y)
			}
		}
	}()
	return ch
}

// BlockIntegrate integrates an image in small batches.
func BlockIntegrate(s *Scene, in Integrator, log log.Logger) <-chan Fragment {
	const blockDim = 32
	numWorkers := runtime.GOMAXPROCS(0)
	cam := s.Camera()
	w, h := cam.ResolutionX(), cam.ResolutionY()
	ch := make(chan Fragment, fragBufferSize)

	// Separate goroutine manages block locations
	locCh := make(chan [2]int)
	go func() {
		defer close(locCh)
		for y := 0; y < h; y += blockDim {
			for x := 0; x < w; x += blockDim {
				locCh <- [2]int{x, y}
			}
		}
	}()

	go func() {
		defer close(ch)
		wg := new(sync.WaitGroup)
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func() {
				for loc := range locCh {
					log.Debugf("Block (%3d, %3d)", loc[0], loc[1])
					for y := loc[1]; y < loc[1]+blockDim && y < h; y++ {
						for x := loc[0]; x < loc[0]+blockDim && x < w; x++ {
							ch <- RenderPixel(s, in, x, y)
						}
					}
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}()
	return ch
}

// WorkerIntegrate integrates an image using a set number of jobs.
func WorkerIntegrate(s *Scene, in Integrator, log log.Logger) <-chan Fragment {
	numWorkers := runtime.GOMAXPROCS(0)
	cam := s.Camera()
	w, h := cam.ResolutionX(), cam.ResolutionY()
	if h < numWorkers {
		return SimpleIntegrate(s, in, log)
	}

	ch := make(chan Fragment, fragBufferSize)
	go func() {
		defer close(ch)
		rowsPerWorker := h / numWorkers
		if h%numWorkers != 0 {
			numWorkers++
		}
		wg := new(sync.WaitGroup)
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(baseY int) {
				for y := baseY; y < baseY+rowsPerWorker && y < h; y++ {
					for x := 0; x < w; x++ {
						ch <- RenderPixel(s, in, x, y)
					}
				}
				wg.Done()
			}(i * rowsPerWorker)
		}
		wg.Wait()
	}()
	return ch
}

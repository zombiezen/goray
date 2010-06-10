//
//  goray/core/integrator.go
//  goray
//
//  Created by Ross Light on 2010-05-29.
//

/* The integrator package provides the interface for rendering methods. */
package integrator

import (
	"goray/core/color"
	"goray/core/ray"
	"goray/core/render"
	"goray/core/scene"
	"goray/core/vector"
)

/* An Integrator renders rays. */
type Integrator interface {
	Preprocess(scene *scene.Scene)
	Integrate(scene *scene.Scene, state *render.State, r ray.Ray) color.AlphaColor
}

/* A SurfaceIntegrator renders rays by casting onto solid objects. */
type SurfaceIntegrator interface {
	Integrator
	SurfaceIntegrator()
}

/* A VolumeIntegrator renders rays by casting onto volumetric regions. */
type VolumeIntegrator interface {
	Integrator
	VolumeIntegrator()
	Transmittance(scene *scene.Scene, state *render.State, r ray.Ray) color.AlphaColor
}

func Render(s *scene.Scene, i Integrator) (img *render.Image) {
	s.Update()
	img = render.NewImage(s.GetCamera().ResolutionX(), s.GetCamera().ResolutionY())
	ch := BlockIntegrate(s, i)
	img.Acquire(ch)
	return
}

func RenderPixel(s *scene.Scene, i Integrator, x, y int) render.Fragment {
	cam := s.GetCamera()
	w, h := cam.ResolutionX(), cam.ResolutionY()
	// Set up state
	state := new(render.State)
	state.Init(nil)
	state.PixelNumber = y*w + x
	state.ScreenPos = vector.New(2.0*float(x)/float(w)-1.0, -2.0*float(y)/float(h)+1.0, 0.0)
	state.Time = 0.0
	// Shoot ray
	r, _ := cam.ShootRay(float(x), float(y), 0, 0)
	// Integrate
	color := i.Integrate(s, state, r)
	return render.Fragment{X: x, Y: y, Color: color}
}

func BlockIntegrate(s *scene.Scene, in Integrator) <-chan render.Fragment {
	const numWorkers = 5
	cam := s.GetCamera()
	w, h := cam.ResolutionX(), cam.ResolutionY()
	ch := make(chan render.Fragment, numWorkers)
	go func() {
		defer close(ch)
		// Set up end signals
		signals := make([]chan bool, numWorkers * 2)
		for i, _ := range signals {
			signals[i] = make(chan bool)
		}
		// Calculate the number of batches needed
		batchCount := w * h / numWorkers
		if w*h%numWorkers != 0 {
			batchCount++
		}

		for batch := 0; batch < batchCount; batch++ {
			// Start new batch
			for i := 0; i < numWorkers; i++ {
				pixelNum := batch*numWorkers + i
				if pixelNum >= w*h {
					break
				}
				go func(pixelNum int, finish chan<- bool) {
					ch <- RenderPixel(s, in, pixelNum%w, pixelNum/w)
					finish <- true
				}(pixelNum, signals[i])
			}
			// Join goroutines
			for i := 0; i < numWorkers; i++ {
				<-signals[i]
			}
		}
	}()
	return ch
}

//
//	goray/core/integrator/integrator.go
//	goray
//
//	Created by Ross Light on 2010-05-29.
//

// The integrator package provides the interface for rendering methods.
package integrator

import (
	"runtime"
	"goray/logging"
	"goray/core/color"
	"goray/core/ray"
	"goray/core/render"
	"goray/core/scene"
	"goray/core/vector"
)

// An Integrator renders rays.
type Integrator interface {
	Preprocess(scene *scene.Scene)
	Integrate(scene *scene.Scene, state *render.State, r ray.DifferentialRay) color.AlphaColor
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
	Transmittance(scene *scene.Scene, state *render.State, r ray.Ray) color.AlphaColor
}

// Render is an easy way of creating an image from a scene.
//
// Render will update the scene, create a new image, and then use one of the
// integration functions to write to the image.
func Render(s *scene.Scene, i Integrator, log logging.Handler) (img *render.Image) {
	s.Update()
	img = render.NewImage(s.GetCamera().ResolutionX(), s.GetCamera().ResolutionY())
	i.Preprocess(s)
	ch := WorkerIntegrate(s, i, log)
	img.Acquire(ch)
	return
}

// RenderPixel creates a fragment for a position in the image.
func RenderPixel(s *scene.Scene, i Integrator, x, y int) render.Fragment {
	cam := s.GetCamera()
	w, h := cam.ResolutionX(), cam.ResolutionY()
	// Set up state
	state := new(render.State)
	state.Init()
	state.PixelNumber = y*w + x
	state.ScreenPos = vector.Vector3D{2.0*float64(x)/float64(w) - 1.0, -2.0*float64(y)/float64(h) + 1.0, 0.0}
	state.Time = 0.0
	// Shoot ray
	r, _ := cam.ShootRay(float64(x), float64(y), 0, 0)
	// Set up differentials
	cRay := ray.DifferentialRay{Ray: r}
	r, _ = cam.ShootRay(float64(x+1), float64(y), 0, 0)
	cRay.FromX = r.From
	cRay.DirX = r.Dir
	r, _ = cam.ShootRay(float64(x), float64(y+1), 0, 0)
	cRay.FromY = r.From
	cRay.DirY = r.Dir
	// Integrate
	color := i.Integrate(s, state, cRay)
	return render.Fragment{X: x, Y: y, Color: color}
}

// SimpleIntegrate integrates an image one pixel at a time.
func SimpleIntegrate(s *scene.Scene, in Integrator, log logging.Handler) <-chan render.Fragment {
	ch := make(chan render.Fragment, 200)
	go func() {
		defer close(ch)
		w, h := s.GetCamera().ResolutionX(), s.GetCamera().ResolutionY()
		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				ch <- RenderPixel(s, in, x, y)
			}
		}
	}()
	return ch
}

// BlockIntegrate integrates an image in small batches.
func BlockIntegrate(s *scene.Scene, in Integrator, log logging.Handler) <-chan render.Fragment {
	const numWorkers = 8
	cam := s.GetCamera()
	w, h := cam.ResolutionX(), cam.ResolutionY()
	ch := make(chan render.Fragment, numWorkers*2)
	go func() {
		defer close(ch)
		// Set up end signals
		signals := make([]chan bool, numWorkers)
		for i, _ := range signals {
			signals[i] = make(chan bool, 1)
		}
		// Calculate the number of batches needed
		batchCount, extras := w*h/numWorkers, w*h%numWorkers
		if extras > 0 {
			batchCount++
		}

		for batch := 0; batch < batchCount; batch++ {
			// If this is the last batch, scale down accordingly.
			if extras > 0 && batch == batchCount-1 {
				signals = signals[0:extras]
			}
			// Start new batch
			for i := 0; i < len(signals); i++ {
				pixelNum := batch*numWorkers + i
				go func(pixelNum int, finish chan<- bool) {
					ch <- RenderPixel(s, in, pixelNum%w, pixelNum/w)
					finish <- true
				}(pixelNum, signals[i])
			}
			// Join goroutines
			for _, sig := range signals {
				<-sig
			}
			logging.Debug(log, "Finished batch %d of %d (%d pixels)", batch+1, batchCount, len(signals))
		}
	}()
	return ch
}

// WorkerIntegrate integrates an image using a set number of jobs.
func WorkerIntegrate(s *scene.Scene, in Integrator, log logging.Handler) <-chan render.Fragment {
	numWorkers := runtime.GOMAXPROCS(0)
	cam := s.GetCamera()
	w, h := cam.ResolutionX(), cam.ResolutionY()
	if h < numWorkers {
		return SimpleIntegrate(s, in, log)
	}

	ch := make(chan render.Fragment, 100)
	go func() {
		defer close(ch)
		// Set up end signals
		signals := make([]chan bool, numWorkers)
		for i, _ := range signals {
			signals[i] = make(chan bool)
		}
		// Start workers
		rowsPerWorker := h / numWorkers
		if h % numWorkers != 0 {
			signals = append(signals, make(chan bool))
		}
		for i, sig := range signals {
			go func(baseY int, finish chan<- bool) {
				for y := baseY; y < baseY+rowsPerWorker && y < h; y++ {
					for x := 0; x < w; x++ {
						ch <- RenderPixel(s, in, x, y)
					}
				}
				finish <- true
			}(i*rowsPerWorker, sig)
		}
		// Join goroutines
		for _, sig := range signals {
			<-sig
		}
	}()
	return ch
}

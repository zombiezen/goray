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
	"sync"
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
	img = render.NewImage(s.Camera().ResolutionX(), s.Camera().ResolutionY())
	i.Preprocess(s)
	ch := BlockIntegrate(s, i, log)
	img.Acquire(ch)
	return
}

// RenderPixel creates a fragment for a position in the image.
func RenderPixel(s *scene.Scene, i Integrator, x, y int) render.Fragment {
	cam := s.Camera()
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
func BlockIntegrate(s *scene.Scene, in Integrator, log logging.Handler) <-chan render.Fragment {
	const blockDim = 32
	numWorkers := runtime.GOMAXPROCS(0)
	cam := s.Camera()
	w, h := cam.ResolutionX(), cam.ResolutionY()
	ch := make(chan render.Fragment, 100)
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
					//logging.Debug(log, "BLOCK (%3d, %3d)", loc[0], loc[1])
					for y := loc[1]; y < loc[1]+blockDim && y < h; y++ {
						for x := loc[0]; x < loc[0]+blockDim && x < w; x++ {
							ch <- RenderPixel(s, in, x, y)
							//logging.VerboseDebug(log, "Rendered (%3d, %3d)", x, y)
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
func WorkerIntegrate(s *scene.Scene, in Integrator, log logging.Handler) <-chan render.Fragment {
	numWorkers := runtime.GOMAXPROCS(0)
	cam := s.Camera()
	w, h := cam.ResolutionX(), cam.ResolutionY()
	if h < numWorkers {
		return SimpleIntegrate(s, in, log)
	}

	ch := make(chan render.Fragment, 100)
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

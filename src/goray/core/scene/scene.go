//
//	goray/core/scene/scene.go
//	goray
//
//	Created by Ross Light on 2010-05-23.
//

// The scene package provides the basic mechanism for establishing an environment to render.
package scene

import (
	"math"
	"os"

	"goray/logging"

	"goray/core/background"
	"goray/core/bound"
	"goray/core/camera"
	"goray/core/intersect"
	"goray/core/light"
	"goray/core/material"
	"goray/core/object"
	"goray/core/primitive"
	"goray/core/ray"
	"goray/core/vector"
	"goray/core/volume"
)

const (
	objectsChanged = 1 << iota
	lightsChanged
	otherChange
)

type changeSet uint

func (c changeSet) IsClear() bool           { return c == 0 }
func (c changeSet) Has(flag changeSet) bool { return c&changeSet(flag) != 0 }
func (c *changeSet) Mark(flag changeSet)    { *c |= changeSet(flag) }
func (c *changeSet) SetAll()                { *c = ^changeSet(0) }
func (c *changeSet) Clear()                 { *c = 0 }

type ObjectID uint

// Scene stores all of the entities that define an environment to render.
// Scene also functions as a high-level API for goray: once you have created a scene, you can create geometry,
// add entities, and render an image.
type Scene struct {
	changes    changeSet
	nextFreeID ObjectID

	log *logging.Logger

	objects    map[ObjectID]object.Object3D
	materials  map[string]material.Material
	volumes    []volume.Region
	lights     []light.Light
	intersect  intersect.Interface
	camera     camera.Camera
	background background.Background
	sceneBound bound.Bound

	aaSamples, aaPasses int
	aaIncSamples        int
	aaThreshold         float64
	doDepth             bool
}

// New creates a new scene.
func New() (s *Scene) {
	s = &Scene{
		log: logging.NewLogger(),

		aaSamples:   1,
		aaPasses:    1,
		aaThreshold: 0.05,

		objects:   make(map[ObjectID]object.Object3D),
		materials: make(map[string]material.Material),
		volumes:   make([]volume.Region, 0),
		lights:    make([]light.Light, 0),

		nextFreeID: 1,
	}

	s.changes.SetAll()
	return s
}

func (s *Scene) Log() *logging.Logger { return s.log }

// AddLight adds a light to the scene.
func (s *Scene) AddLight(l light.Light) (err os.Error) {
	if l == nil {
		return os.NewError("Attempted to insert nil light")
	}
	s.lights = append(s.lights, l)
	s.changes.Mark(lightsChanged)
	return
}

// Lights returns all of the lights added to the scene.
func (s *Scene) Lights() []light.Light { return s.lights }

// AddMaterial adds a material to the scene.
func (s *Scene) AddMaterial(name string, m material.Material) (err os.Error) {
	return os.NewError("We don't support named materials yet")
}

// AddObject adds a three-dimensional object to the scene.
func (s *Scene) AddObject(obj object.Object3D) (id ObjectID, err os.Error) {
	id = s.nextFreeID
	if _, found := s.objects[id]; found {
		err = os.NewError("Internal error: allocated ID is already in use")
		return
	}
	// Add into map
	s.objects[id] = obj
	s.nextFreeID++
	s.changes.Mark(objectsChanged)
	return
}

// GetObject retrieves the object with a given ID.
func (s *Scene) GetObject(id ObjectID) (obj object.Object3D, found bool) {
	obj, found = s.objects[id]
	return
}

// AddVolumeRegion adds a volumetric effect to the scene.
func (s *Scene) AddVolumeRegion(vr volume.Region) { s.volumes = append(s.volumes, vr) }

// Camera returns the scene's current camera.
func (s *Scene) Camera() camera.Camera { return s.camera }

// SetCamera changes the scene's current camera.
func (s *Scene) SetCamera(cam camera.Camera) { s.camera = cam }

// Background returns the scene's current background.
func (s *Scene) Background() background.Background { return s.background }

// SetBackground changes the scene's current background.
func (s *Scene) SetBackground(bg background.Background) { s.background = bg }

// SetAntialiasing changes the parameters for antialiasing.
func (s *Scene) SetAntialiasing(numSamples, numPasses, incSamples int, threshold float64) {
	if numSamples < 1 {
		numSamples = 1
	}
	s.aaSamples = numSamples
	s.aaPasses = numPasses
	if incSamples > 0 {
		s.aaIncSamples = incSamples
	} else {
		s.aaIncSamples = s.aaSamples
	}
	s.aaThreshold = threshold
}

// Bound returns a bounding box that contains every object in the scene.
func (s *Scene) Bound() bound.Bound { return s.sceneBound }

func (s *Scene) DoDepth() bool { return s.doDepth }

// Intersect returns the surface point that intersects with the given ray.
func (s *Scene) Intersect(r ray.Ray, dist float64) primitive.Collision {
	// Determine distance to check
	if dist < 0 {
		dist = r.TMax
		if dist < 0 {
			dist = math.Inf(1)
		}
	}
	// Perform low-level intersection
	if s.intersect == nil {
		s.log.Warning("Intersect called without an Update")
		return primitive.Collision{}
	}
	return s.intersect.Intersect(r, dist)
}

// Shadowed returns whether a ray will cast a shadow.
func (s *Scene) Shadowed(r ray.Ray, dist float64) bool {
	if s.intersect == nil {
		s.log.Warning("Shadowed called without an Update")
		return false
	}
	r.From = vector.Add(r.From, vector.ScalarMul(r.Dir, r.TMin))
	if r.TMax >= 0 {
		dist = r.TMax - 2*r.TMin
	}
	return s.intersect.Shadowed(r, dist)
}

// Update causes the scene state to prepare for rendering.
// This is a potentially expensive operation.  It will be called automatically before a Render.
func (s *Scene) Update() (err os.Error) {
	if s.changes.IsClear() {
		// Already up-to-date.
		return
	}
	if s.camera == nil {
		return os.NewError("Scene has no camera")
	}
	s.log.Debug("Performing scene update...")

	if s.changes.Has(objectsChanged) {
		// We've changed the scene's geometry.  We need to rebuild the intersection scheme.
		s.intersect = nil
		// Collect primitives
		prims := make([]primitive.Primitive, 0, len(s.objects))
		for _, obj := range s.objects {
			prims = append(prims, obj.Primitives()...)
		}
		s.log.Debug("Geometry collected, %d primitives", len(prims))
		// Do partition building
		if len(prims) > 0 {
			s.log.Debug("Building kd-tree...")
			s.intersect = intersect.NewKD(prims, s.log)
			s.sceneBound = s.intersect.Bound()
			s.log.Debug("Built kd-tree")
		}
	}

	if s.changes.Has(lightsChanged) {
		for _, li := range s.lights {
			li.SetScene(s)
		}
		if s.background != nil {
			bgLight := s.background.Light()
			if bgLight != nil {
				bgLight.SetScene(s)
			}
		}
		s.log.Debug("Set up lights")
	}

	s.changes.Clear()
	return
}

//
//  goray/core/scene.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

/* The scene package provides the basic mechanism for establishing an environment to render. */
package scene

import (
	container "container/vector"
	"os"

	"goray/fmath"
	"goray/logging"

	"goray/core/background"
	"goray/core/bound"
	"goray/core/camera"
	"goray/core/light"
	"goray/core/material"
	"goray/core/object"
	"goray/core/partition"
	"goray/core/primitive"
	"goray/core/ray"
	"goray/core/render"
	"goray/core/surface"
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

/*
   Scene stores all of the entities that define an environment to render.
   Scene also functions as a high-level API for goray: once you have created a scene, you can create geometry,
   add entities, and render an image.
*/
type Scene struct {
	changes    changeSet
	nextFreeID ObjectID

	log *logging.Logger

	objects     map[ObjectID]object.Object3D
	materials   map[string]material.Material
	volumes     container.Vector
	lights      container.Vector
	partitioner partition.Partitioner
	camera      camera.Camera
	background  background.Background
	sceneBound  *bound.Bound

	aaSamples, aaPasses int
	aaIncSamples        int
	aaThreshold         float
	doDepth             bool
}

/* New creates a new scene */
func New() *Scene {
	s := new(Scene)
	s.log = logging.NewLogger()

	s.aaSamples = 1
	s.aaPasses = 1
	s.aaThreshold = 0.05

	s.objects = make(map[ObjectID]object.Object3D)
	s.materials = make(map[string]material.Material)
	s.volumes = make(container.Vector, 0)
	s.lights = make(container.Vector, 0)

	s.changes.SetAll()
	s.nextFreeID = 1
	return s
}

func (s *Scene) GetLog() *logging.Logger { return s.log }

/* AddLight adds a light to the scene. */
func (s *Scene) AddLight(l light.Light) (err os.Error) {
	if l == nil {
		return os.NewError("Attempted to insert nil light")
	}
	s.lights.Push(l)
	s.changes.Mark(lightsChanged)
	return
}

/* GetLights returns all of the lights added to the scene. */
func (s *Scene) GetLights() []light.Light {
	temp := make([]light.Light, s.lights.Len())
	it := s.lights.Iter()
	for i, val := 0, <-it; !closed(it); i, val = i+1, <-it {
		temp[i] = val.(light.Light)
	}
	return temp
}

/* AddMaterial adds a material to the scene. */
func (s *Scene) AddMaterial(name string, m material.Material) (err os.Error) {
	return os.NewError("We don't support named materials yet")
}

/* AddObject adds a three-dimensional object to the scene. */
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

/* GetObject retrieves the object with a given ID. */
func (s *Scene) GetObject(id ObjectID) (obj object.Object3D, found bool) {
	obj, found = s.objects[id]
	return
}

/* AddVolumeRegion adds a volumetric effect to the scene. */
func (s *Scene) AddVolumeRegion(vr volume.Region) { s.volumes.Push(vr) }

/* GetCamera returns the scene's current camera. */
func (s *Scene) GetCamera() camera.Camera { return s.camera }

/* SetCamera changes the scene's current camera. */
func (s *Scene) SetCamera(cam camera.Camera) { s.camera = cam }

/* GetBackground returns the scene's current background. */
func (s *Scene) GetBackground() background.Background { return s.background }

/* SetBackground changes the scene's current background. */
func (s *Scene) SetBackground(bg background.Background) { s.background = bg }

/* SetAntialiasing changes the parameters for antialiasing. */
func (s *Scene) SetAntialiasing(numSamples, numPasses, incSamples int, threshold float) {
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

/* GetSceneBound returns a bounding box that contains every object in the scene. */
func (s *Scene) GetSceneBound() *bound.Bound { return s.sceneBound }

func (s *Scene) GetDoDepth() bool { return s.doDepth }

/* Intersect returns the surface point that intersects with the given ray. */
func (s *Scene) Intersect(r ray.Ray) (coll primitive.Collision, sp surface.Point, err os.Error) {
	dist := r.TMax()
	if r.TMax() < 0 {
		dist = fmath.Inf
	}
	// Intersect with partitioner
	if s.partitioner == nil {
		err = os.NewError("Partition map has not been built")
		return
	}
	coll = s.partitioner.Intersect(r, dist)
	if !coll.Hit() {
		return
	}
	sp = coll.Primitive.GetSurface(coll)
	sp.Primitive = coll.Primitive
	return
}

/* IsShadowed returns whether a ray will cast a shadow. */
func (s *Scene) IsShadowed(state *render.State, r ray.Ray) bool {
	if s.partitioner == nil {
		return false
	}
	r.SetFrom(vector.Add(r.From(), vector.ScalarMul(r.Dir(), r.TMin())))
	r.SetTime(state.Time)
	dist := fmath.Inf
	if r.TMax() >= 0 {
		dist = r.TMax() - 2*r.TMin()
	}
	coll := s.partitioner.IntersectS(r, dist)
	return coll.Hit()
}

/*
   Update causes the scene state to prepare for rendering.
   This is a potentially expensive operation.  It will be called automatically before a Render.
*/
func (s *Scene) Update() (err os.Error) {
	if s.camera == nil {
		return os.NewError("Scene has no camera")
	}

	s.log.Debug("Performing scene update...")

	if s.changes.Has(objectsChanged) {
		// We've changed the scene's geometry.  We need to rebuild the partitioner.
		s.partitioner = nil
		// Collect primitives
		var prims []primitive.Primitive
		{
			nPrims := 0
			primLists := make([][]primitive.Primitive, len(s.objects))

			i := 0
			for _, obj := range s.objects {
				primLists[i] = obj.GetPrimitives()
				nPrims += len(primLists[i])
				i++
			}

			prims = make([]primitive.Primitive, nPrims)
			pos := 0
			for _, pl := range primLists {
				copy(prims[pos:], pl)
				pos += len(pl)
			}
		}
		s.log.Debug("Geometry collected, %d primitives", len(prims))
		// Do partition building
		if len(prims) > 0 {
			s.log.Debug("Building kd-tree...")
			s.partitioner = partition.NewKD(prims, s.log)
			s.sceneBound = s.partitioner.GetBound()
			s.log.Debug("Built kd-tree")
		}
	}

	s.lights.Do(func(obj interface{}) {
		li := obj.(light.Light)
		li.SetScene(s)
	})

	if s.background != nil {
		bgLight := s.background.GetLight()
		if bgLight != nil {
			bgLight.SetScene(s)
		}
	}

	s.log.Debug("Set up lights")

	s.changes.Clear()
	return
}

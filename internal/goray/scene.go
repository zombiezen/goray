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
	"errors"
	"math"

	"bitbucket.org/zombiezen/math3/vec64"
	"zombiezen.com/go/goray/internal/bound"
	"zombiezen.com/go/goray/internal/color"
	"zombiezen.com/go/goray/internal/log"
)

const (
	sceneObjectsChanged = 1 << iota
	sceneLightsChanged
	sceneOtherChange
)

type sceneChangeSet uint

func (c sceneChangeSet) IsClear() bool                { return c == 0 }
func (c sceneChangeSet) Has(flag sceneChangeSet) bool { return c&sceneChangeSet(flag) != 0 }

func (c *sceneChangeSet) Mark(flag sceneChangeSet) { *c |= sceneChangeSet(flag) }
func (c *sceneChangeSet) SetAll()                  { *c = ^sceneChangeSet(0) }
func (c *sceneChangeSet) Clear()                   { *c = 0 }

type ObjectID uint

// Intersecter defines a type that can detect goray collisions.
//
// For most cases, this will involve an algorithm that partitions the scene to make these operations faster.
type Intersecter interface {
	// Intersect determines the goray that a goray collides with.
	Intersect(r Ray, dist float64) Collision

	// Shadowed checks whether a goray collides with any primitives (for shadow-detection).
	Shadowed(r Ray, dist float64) bool

	// TransparentShadow computes the color of a transparent shadow after being
	// filtered by the objects in the scene.  The resulting color can be
	// multiplied by the light color to determine the color of the shadow.  The
	// hit return value is set when an opaque goray is encountered or the
	// maximum depth is exceeded.
	TransparentShadow(state *RenderState, r Ray, maxDepth int, dist float64) (filt color.Color, hit bool)

	// Bound returns a bounding box that contains all of the primitives that the intersecter knows about.
	Bound() bound.Bound
}

type IntersecterBuilder func([]Primitive, log.Logger) Intersecter

// Scene stores all of the entities that define an environment to render.
// Scene also functions as a high-level API for goray: once you have created a scene, you can create geometry,
// add entities, and render an image.
type Scene struct {
	changes    sceneChangeSet
	nextFreeID ObjectID

	log log.Logger

	objects    map[ObjectID]Object3D
	materials  map[string]Material
	volumes    []VolumeRegion
	lights     []Light
	camera     Camera
	background Background

	intersecter        Intersecter
	intersecterBuilder IntersecterBuilder
	sceneBound         bound.Bound

	aaSamples, aaPasses int
	aaIncSamples        int
	aaThreshold         float64
	doDepth             bool
}

// NewScene creates a new scene.
func NewScene(ib IntersecterBuilder, l log.Logger) (s *Scene) {
	s = &Scene{
		log: l,

		aaSamples:   1,
		aaPasses:    1,
		aaThreshold: 0.05,

		objects:   make(map[ObjectID]Object3D),
		materials: make(map[string]Material),
		volumes:   make([]VolumeRegion, 0),
		lights:    make([]Light, 0),

		intersecterBuilder: ib,

		nextFreeID: 1,
	}

	s.changes.SetAll()
	return s
}

// AddLight adds a light to the scene.
func (s *Scene) AddLight(l Light) (err error) {
	if l == nil {
		return errors.New("Attempted to insert nil light")
	}
	s.lights = append(s.lights, l)
	s.changes.Mark(sceneLightsChanged)
	return
}

// Lights returns all of the lights added to the scene.
func (s *Scene) Lights() []Light { return s.lights }

// AddMaterial adds a material to the scene.
func (s *Scene) AddMaterial(name string, m Material) (err error) {
	return errors.New("We don't support named materials yet")
}

// AddObject adds a three-dimensional object to the scene.
func (s *Scene) AddObject(obj Object3D) (id ObjectID, err error) {
	id = s.nextFreeID
	if _, found := s.objects[id]; found {
		err = errors.New("Internal error: allocated ID is already in use")
		return
	}

	// Add into map
	s.objects[id] = obj
	s.nextFreeID++
	s.changes.Mark(sceneObjectsChanged)
	return
}

// GetObject retrieves the object with a given ID.
func (s *Scene) GetObject(id ObjectID) (obj Object3D, found bool) {
	obj, found = s.objects[id]
	return
}

// AddVolumeRegion adds a volumetric effect to the scene.
func (s *Scene) AddVolumeRegion(vr VolumeRegion) {
	s.volumes = append(s.volumes, vr)
}

// Camera returns the scene's current camera.
func (s *Scene) Camera() Camera {
	return s.camera
}

// SetCamera changes the scene's current camera.
func (s *Scene) SetCamera(cam Camera) {
	s.camera = cam
}

// Background returns the scene's current background.
func (s *Scene) Background() Background {
	return s.background
}

// SetBackground changes the scene's current background.
func (s *Scene) SetBackground(bg Background) {
	s.background = bg
}

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
func (s *Scene) Bound() bound.Bound {
	return s.sceneBound
}

func (s *Scene) DoDepth() bool {
	return s.doDepth
}

// Intersect returns the surface point that intersects with the given ray.
func (s *Scene) Intersect(r Ray, dist float64) Collision {
	// Determine distance to check
	if dist < 0 {
		dist = r.TMax
		if dist < 0 {
			dist = math.Inf(1)
		}
	}

	// Perform low-level intersection
	if s.intersecter == nil {
		s.log.Warningf("Intersect called without an Update")
		return Collision{}
	}
	return s.intersecter.Intersect(r, dist)
}

// Shadowed returns whether a ray will cast a shadow.
func (s *Scene) Shadowed(r Ray, dist float64) bool {
	if s.intersecter == nil {
		s.log.Warningf("Shadowed called without an Update")
		return false
	}
	r.From = vec64.Add(r.From, r.Dir.Scale(r.TMin))
	if r.TMax >= 0 {
		dist = r.TMax - 2*r.TMin
	}
	return s.intersecter.Shadowed(r, dist)
}

// Update causes the scene state to prepare for rendering.
// This is a potentially expensive operation.  It will be called automatically before a Render.
func (s *Scene) Update() (err error) {
	if s.changes.IsClear() {
		// Already up-to-date.
		return
	}
	if s.camera == nil {
		return errors.New("Scene has no camera")
	}
	s.log.Debugf("Performing scene update...")

	if s.changes.Has(sceneObjectsChanged) {
		// We've changed the scene's geometry.  We need to rebuild the intersection scheme.
		s.intersecter = nil

		// Collect primitives
		prims := make([]Primitive, 0, len(s.objects))
		for _, obj := range s.objects {
			prims = append(prims, obj.Primitives()...)
		}
		s.log.Debugf("Geometry collected, %d primitives", len(prims))

		// Do partition building
		if len(prims) > 0 {
			s.intersecter = s.intersecterBuilder(prims, s.log)
			s.sceneBound = s.intersecter.Bound()
		}
	}

	if s.changes.Has(sceneLightsChanged) {
		for _, li := range s.lights {
			li.SetScene(s)
		}
		if s.background != nil {
			bgLight := s.background.Light()
			if bgLight != nil {
				bgLight.SetScene(s)
			}
		}
		s.log.Debugf("Set up lights")
	}

	s.changes.Clear()
	return
}

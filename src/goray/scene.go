//
//  goray/scene.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

/* The goray/scene package provides the basic mechanism for establishing an environment to render. */
package scene

import (
	vecarray "container/vector"
	"os"
	"./fmath"
	"./stack"
)

import (
	"./goray/background"
	"./goray/bound"
	"./goray/camera"
	"./goray/integrator"
	"./goray/light"
	"./goray/material"
	"./goray/object"
	"./goray/partition"
	"./goray/primitive"
	"./goray/ray"
	"./goray/render"
	"./goray/surface"
	"./goray/vector"
	"./goray/vmap"
	"./goray/volume"
)

const (
	stateReady = iota
	stateGeometry
	stateObject
	stateVmap
)

const (
	changeGeom = 1 << iota
	changeLight
	changeOther

	changeNone = 0
	changeAll  = changeGeom | changeLight | changeOther
)

type objData struct {
	points   []vector.Vector3D
	normals  []vector.Vector3D
	dataType int
}

type ObjectID uint

/*
   Scene stores all of the entities that define an environment to render.
   Scene also functions as a high-level API for goray: once you have created a scene, you can create geometry,
   add entities, and render an image.
*/
type Scene struct {
	state struct {
		stack       *stack.Stack
		changes     uint
		nextFreeID  ObjectID
		currObj     *objData
		currVmap    *vmap.VMap
		orco        bool
		smoothAngle float
	}

	objects        map[ObjectID]object.Object3D
	meshes         map[ObjectID]objData
	materials      map[string]material.Material
	volumes        vecarray.Vector
	lights         vecarray.Vector
	vmaps          map[int]int
	tree           partition.Partitioner
	camera         camera.Camera
	background     background.Background
	volIntegrator  integrator.VolumeIntegrator
	surfIntegrator integrator.SurfaceIntegrator
	sceneBound     *bound.Bound

	aaSamples, aaPasses int
	aaIncSamples        int
	aaThreshold         float
	mode                int
	doDepth             bool
}

/* New creates a new scene */
func New() *Scene {
	s := new(Scene)
	s.aaSamples = 1
	s.aaPasses = 1
	s.aaThreshold = 0.05

	s.objects = make(map[ObjectID]object.Object3D)
	s.meshes = make(map[ObjectID]objData)
	s.materials = make(map[string]material.Material)
	s.volumes = make(vecarray.Vector, 0)
	s.lights = make(vecarray.Vector, 0)
	s.vmaps = make(map[int]int)

	s.state.changes = changeAll
	s.state.stack = stack.New()
	s.state.stack.Push(stateReady)
	s.state.nextFreeID = 1
	s.state.currObj = nil
	return s
}

func (s *Scene) currState() int {
	cs, _ := s.state.stack.Top()
	return cs.(int)
}

/* StartGeometry puts the scene in geometry creation mode. */
func (s *Scene) StartGeometry() (err os.Error) {
	if s.currState() != stateReady {
		return os.NewError("Scene asked to start geometry in wrong mode")
	}
	s.state.stack.Push(stateGeometry)
	return
}

/* EndGeometry finishes geometry creation mode. */
func (s *Scene) EndGeometry() (err os.Error) {
	if s.currState() != stateGeometry {
		return os.NewError("Scene asked to end geometry in wrong mode")
	}
	s.state.stack.Pop()
	return
}

//func (s *Scene) StartTriMesh(vertices, triangles int, hasOrco, hasUV bool, meshType int) (bool, ObjectID) {
//
//}
//
//func (s *Scene) EndTriMesh() bool {
//
//}
//
//func (s *Scene) AddVertex(p Vector3D) int {
//
//}
//
//func (s *Scene) AddTriangle(a, b, c int, mat Material) bool {
//
//}
//
//func (s *Scene) AddUVTriangle(a, b, c int, uvA, uvB, uvC int, mat Material) bool {
//
//}

/* AddLight adds a light to the scene. */
func (s *Scene) AddLight(l light.Light) (err os.Error) {
	if l == nil {
		return os.NewError("Attempted to insert nil light")
	}
	s.lights.Push(l)
	s.state.changes |= changeLight
	return
}

/* AddMaterial adds a material to the scene. */
func (s *Scene) AddMaterial(name string, m material.Material) (err os.Error) {
	return os.NewError("We don't support named materials yet")
}

/* AddObject adds a three-dimensional object to the scene. */
func (s *Scene) AddObject(obj object.Object3D) (id ObjectID, err os.Error) {
	id = s.state.nextFreeID
	if _, found := s.objects[id]; found {
		err = os.NewError("Internal error: allocated ID is already in use")
		return
	}
	// Add into map
	s.objects[id] = obj
	s.state.nextFreeID++
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

/* SetSurfaceIntegrator changes the integrator used for evaluating surfaces. */
func (s *Scene) SetSurfaceIntegrator(i integrator.SurfaceIntegrator) {
	s.surfIntegrator = i
	s.surfIntegrator.SetScene(s)
	s.state.changes |= changeOther
}

/* SetVolumeIntegrator changes the integrator used for evaluating volumes. */
func (s *Scene) SetVolumeIntegrator(i integrator.VolumeIntegrator) {
	s.volIntegrator = i
	s.volIntegrator.SetScene(s)
	s.state.changes |= changeOther
}

/* GetSceneBound returns a bounding box that contains every object in the scene. */
func (s *Scene) GetSceneBound() *bound.Bound { return s.sceneBound }

func (s *Scene) GetDoDepth() bool { return s.doDepth }

/* Intersect returns the surface point that intersects with the given ray. */
func (s *Scene) Intersect(r ray.Ray) (sp surface.Point, hit bool, err os.Error) {
	dist := r.TMax()
	if r.TMax() < 0 {
		dist = fmath.Inf
	}
	// Intersect with tree
	if s.tree == nil {
		err = os.NewError("Partition map has not been built")
		return
	}
	coll := s.tree.Intersect(r, dist)
	if !coll.Hit() {
		return
	}
	sp = coll.Primitive.GetSurface(coll)
	sp.Primitive = coll.Primitive
	return
}

/* IsShadowed returns whether a ray will cast a shadow. */
func (s *Scene) IsShadowed(state *render.State, r ray.Ray) bool {
	if s.tree == nil {
		return false
	}
	r.SetFrom(vector.Add(r.From(), vector.ScalarMul(r.Dir(), r.TMin())))
	r.SetTime(state.Time)
	dist := fmath.Inf
	if r.TMax() >= 0 {
		dist = r.TMax() - 2*r.TMin()
	}
	coll := s.tree.IntersectS(r, dist)
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

	if s.state.changes&changeGeom != 0 {
		// We've changed the scene's geometry.  We need to rebuild the tree.
		s.tree = nil
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
		// Do tree building
		if len(prims) > 0 {
			//s.tree = kdtree.New(prims, -1, 1, 0.8, 0.33)
			s.tree = partition.NewSimple(prims)
			s.sceneBound = s.tree.GetBound()
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

	if s.surfIntegrator == nil {
		return os.NewError("Scene has no surface integrator")
	}

	if s.state.changes != changeNone {
		if err = s.surfIntegrator.Preprocess(); err != nil {
			return
		}
		if s.volIntegrator != nil {
			if err = s.volIntegrator.Preprocess(); err != nil {
				return
			}
		}
	}
	s.state.changes = changeNone
	return
}

/* Render creates an image of the scene. */
func (s *Scene) Render() (img *render.Image, err os.Error) {
	err = s.Update()
	if err != nil {
		return
	}

	img = render.NewImage(s.camera.ResolutionX(), s.camera.ResolutionY())
	ch := s.surfIntegrator.Render()
	img.Acquire(ch)
	return
}

//
//  goray/scene.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

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
	volumes        *vecarray.Vector
	lights         *vecarray.Vector
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

func NewScene() *Scene {
	s := new(Scene)
	s.aaSamples = 1
	s.aaPasses = 1
	s.aaThreshold = 0.05
	s.state.changes = changeAll
	s.state.stack.Push(stateReady)
	s.state.nextFreeID = 1
	s.state.currObj = nil
	return s
}

func (s *Scene) currState() int {
	cs, _ := s.state.stack.Top()
	return cs.(int)
}

func (s *Scene) StartGeometry() (err os.Error) {
	if s.currState() != stateReady {
		return os.NewError("Scene asked to start geometry in wrong mode")
	}
	s.state.stack.Push(stateGeometry)
	return
}

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

func (s *Scene) AddLight(l light.Light) (err os.Error) {
	if l == nil {
		return os.NewError("Attempted to insert nil light")
	}
	s.lights.Push(l)
	s.state.changes |= changeLight
	return
}

func (s *Scene) AddMaterial(name string, m material.Material) (err os.Error) {
	return os.NewError("We don't support named materials yet")
}

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

func (s *Scene) GetObject(id ObjectID) (obj object.Object3D, found bool) {
	obj, found = s.objects[id]
	return
}

func (s *Scene) AddVolumeRegion(vr volume.Region) { s.volumes.Push(vr) }

func (s *Scene) GetCamera() camera.Camera    { return s.camera }
func (s *Scene) SetCamera(cam camera.Camera) { s.camera = cam }

func (s *Scene) GetBackground() background.Background   { return s.background }
func (s *Scene) SetBackground(bg background.Background) { s.background = bg }

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

func (s *Scene) SetSurfaceIntegrator(i integrator.SurfaceIntegrator) {
	s.surfIntegrator = i
	s.surfIntegrator.SetScene(s)
	s.state.changes |= changeOther
}

func (s *Scene) SetVolumeIntegrator(i integrator.VolumeIntegrator) {
	s.volIntegrator = i
	s.volIntegrator.SetScene(s)
	s.state.changes |= changeOther
}

func (s *Scene) GetSceneBound() *bound.Bound { return s.sceneBound }

func (s *Scene) GetDoDepth() bool { return s.doDepth }

func (s *Scene) Intersect(r ray.Ray) (sp surface.Point, hit bool, err os.Error) {
	dist := r.TMax
	if r.TMax < 0 {
		dist = fmath.Inf
	}
	_ = dist // for now
	// Intersect with tree
	if s.tree == nil {
		err = os.NewError("Partition map has not been built")
		return
	}
	hit, hitprim, z := s.tree.Intersect(r, dist)
	if !hit {
		return
	}
	h := vector.Add(r.From, vector.ScalarMul(r.Dir, z))
	sp = hitprim.GetSurface(h)
	sp.Primitive = hitprim
	return
}

func (s *Scene) IsShadowed(state *render.State, r ray.Ray) bool {
	if s.tree == nil {
		return false
	}
	r.From = vector.Add(r.From, vector.ScalarMul(r.Dir, r.TMin))
	r.Time = state.Time
	dist := fmath.Inf
	if r.TMax >= 0 {
		dist = r.TMax - 2*r.TMin
	}
	hit, _ := s.tree.IntersectS(r, dist)
	return hit
}

// Update scene state to prepare for rendering
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

			for i, obj := range s.objects {
				primLists[i] = obj.GetPrimitives()
				nPrims += len(primLists[i])
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
		li.Init(s)
	})

	if s.background != nil {
		bgLight := s.background.GetLight()
		if bgLight != nil {
			bgLight.Init(s)
		}
	}

	if s.surfIntegrator == nil {
		return os.NewError("Scene has no surface integrator")
	}

	if s.state.changes != changeNone {
		if err = s.surfIntegrator.Preprocess(); err != nil {
			return
		}
		if err = s.volIntegrator.Preprocess(); err != nil {
			return
		}
	}
	s.state.changes = changeNone
	return
}

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

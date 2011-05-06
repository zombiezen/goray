//
//	goray/std/all.go
//	goray
//
//	Created by Ross Light on 2011-04-05.
//

/*
	The all package does not provide any functionality, but merely imports all
	the packages in the std directory.  This, in turn, registers all the
	packages with yamlscene.  Typical usage:

		import _ "goray/std/all"
*/
package all

import (
	_ "goray/std/cameras/ortho"
	_ "goray/std/cameras/perspective"
	_ "goray/std/integrators/directlight"
	_ "goray/std/lights/point"
	_ "goray/std/lights/spot"
	_ "goray/std/materials/debug"
	_ "goray/std/materials/shinydiffuse"
	_ "goray/std/objects/mesh"
	_ "goray/std/shaders/texmap"
	_ "goray/std/textures/image"
)

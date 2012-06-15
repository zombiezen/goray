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

/*
	Package std does not provide any functionality, but merely imports all
	the packages in the std directory.  This, in turn, registers all the
	packages with yamlscene.  Typical usage:

		import _ "goray/std"
*/
package std

import (
	_ "bitbucket.org/zombiezen/goray/std/cameras/ortho"
	_ "bitbucket.org/zombiezen/goray/std/cameras/perspective"
	_ "bitbucket.org/zombiezen/goray/std/integrators/directlight"
	_ "bitbucket.org/zombiezen/goray/std/lights/point"
	_ "bitbucket.org/zombiezen/goray/std/lights/spot"
	_ "bitbucket.org/zombiezen/goray/std/materials/debug"
	_ "bitbucket.org/zombiezen/goray/std/materials/shinydiffuse"
	_ "bitbucket.org/zombiezen/goray/std/shaders/texmap"
	_ "bitbucket.org/zombiezen/goray/std/textures/image"
)

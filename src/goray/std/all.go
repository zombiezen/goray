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
	Package all does not provide any functionality, but merely imports all
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

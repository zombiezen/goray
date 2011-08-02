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

package job

import (
	"io"
	"os"

	"image"
	"image/jpeg"
	"image/png"
)

// A Format holds information about how to encode an job's image.
type Format struct {
	Extension string
	Encode    func(io.Writer, image.Image) os.Error
}

const DefaultFormat = "png"

var FormatMap = map[string]Format{
	"png":  Format{".png", png.Encode},
	"jpeg": Format{".jpg", func(w io.Writer, i image.Image) os.Error { return jpeg.Encode(w, i, nil) }},
}

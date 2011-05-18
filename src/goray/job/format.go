//
//  goray/job/format.go
//  goray
//
//  Created by Ross Light on 2011-05-18.
//

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

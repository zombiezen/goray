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

package textures

import (
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	slashpath "path"
	"path/filepath"

	"zombiezen.com/go/goray/internal/goray"
)

// ImageLoader is an interface for retrieving images with a name.
type ImageLoader interface {
	LoadImage(name string) (img *goray.Image, err error)
}

// ImageLoaderFunc uses a function to perform loads.
type ImageLoaderFunc func(string) (*goray.Image, error)

func (f ImageLoaderFunc) LoadImage(name string) (img *goray.Image, err error) {
	return f(name)
}

type fileImageLoader struct {
	Base  string
	Clean bool
}

func (l *fileImageLoader) LoadImage(name string) (*goray.Image, error) {
	if name == "" {
		return nil, errors.New("name must not be empty")
	}
	if l.Clean {
		name = slashpath.Clean("/" + name)
	}
	path := filepath.FromSlash(name)
	if l.Clean || name[0] != '/' {
		path = filepath.Join(l.Base, path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	i, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return goray.NewGoImage(i), nil
}

// NewImageLoader creates an image loader that defaults to the given directory.
// Users of the loader can access anything in local storage.
func NewImageLoader(base string) ImageLoader {
	return &fileImageLoader{Base: base}
}

// NewImageLoaderDirectory creates an image loader that is rooted at a given
// directory.  Users of the loader will not directly be able to access anything
// outside the directory, but symlinks inside the directory will be followed.
func NewImageLoaderDirectory(base string) ImageLoader {
	return &fileImageLoader{Base: base, Clean: true}
}

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

// Package fileloader provides two image texture loaders for the local filesystem.
package fileloader

import (
	"os"

	goimage "image"
	_ "image/jpeg"
	_ "image/png"

	slashpath "path"
	"path/filepath"

	"goray"
	"goray/std/textures/image"
)

func openImage(fspath string) (img *goray.Image, err os.Error) {
	f, err := os.Open(fspath)
	if err != nil {
		return
	}
	defer f.Close()
	i, _, err := goimage.Decode(f)
	if err != nil {
		return
	}
	return goray.NewGoImage(i), nil
}

// New creates an image loader that is rooted at a given directory.
// Users of the loader will not directly be able to access anything outside the directory, but symlinks inside the directory will be followed.
func New(base string) image.Loader {
	return image.LoaderFunc(func(name string) (img *goray.Image, err os.Error) {
		if name == "" {
			return nil, os.NewError("name must not be empty")
		}
		name = slashpath.Clean("/" + name)
		return openImage(slashpath.Join(base, filepath.FromSlash(name)))
	})
}

// NewFull creates an image loader that defaults to the given directory.
// Users of the loader can access anything in local storage.
func NewFull(base string) image.Loader {
	return image.LoaderFunc(func(name string) (img *goray.Image, err os.Error) {
		if name == "" {
			return nil, os.NewError("name must not be empty")
		}
		var p string
		if name[0] == '/' {
			p = filepath.FromSlash(name)
		} else {
			p = slashpath.Join(base, filepath.FromSlash(name))
		}
		return openImage(p)
	})
}

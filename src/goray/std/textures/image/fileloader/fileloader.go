//
//	goray/std/textures/image/fileloader/fileloader.go
//	goray
//
//	Created by Ross Light on 2011-04-30.
//

// Package fileloader provides two image texture loaders for the local filesystem.
package fileloader

import (
	"os"

	goimage "image"
	_ "image/jpeg"
	_ "image/png"

	slashpath "path"
	"path/filepath"

	"goray/core/render"
	"goray/std/textures/image"
)

func openImage(fspath string) (img *render.Image, err os.Error) {
	f, err := os.Open(fspath)
	if err != nil {
		return
	}
	defer f.Close()
	i, _, err := goimage.Decode(f)
	if err != nil {
		return
	}
	return render.NewGoImage(i), nil
}

// New creates an image loader that is rooted at a given directory.
// Users of the loader will not directly be able to access anything outside the directory, but symlinks inside the directory will be followed.
func New(base string) image.Loader {
	return image.LoaderFunc(func(name string) (img *render.Image, err os.Error) {
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
	return image.LoaderFunc(func(name string) (img *render.Image, err os.Error) {
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

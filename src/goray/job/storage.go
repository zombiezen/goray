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
	"path/filepath"
)

// Storage defines an interface for a manager to store and retrieve render job results.
type Storage interface {
	OpenReader(job *Job) (rc io.ReadCloser, err os.Error)
	OpenWriter(job *Job) (wc io.WriteCloser, err os.Error)
}

type fileStorage struct {
	Root string
}

// NewFileStorage creates a Storage based on a root directory.
func NewFileStorage(rootDir string) (storage Storage, err os.Error) {
	err = os.Mkdir(rootDir, 0777)
	if err != nil {
		if pe, ok := err.(*os.PathError); !ok || pe.Error != os.EEXIST {
			return
		}
		err = nil
	}
	storage = fileStorage{Root: rootDir}
	return
}

func (fs fileStorage) path(j *Job) string {
	return filepath.Join(fs.Root, j.Name+".png")
}

func (fs fileStorage) OpenReader(j *Job) (rc io.ReadCloser, err os.Error) {
	rc, err = os.Open(fs.path(j))
	return
}

func (fs fileStorage) OpenWriter(j *Job) (wc io.WriteCloser, err os.Error) {
	wc, err = os.Create(fs.path(j))
	return
}

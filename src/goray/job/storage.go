//
//  goray/job/storage.go
//  goray
//
//  Created by Ross Light on 2011-03-14.
//

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
	rc, err = os.Open(fs.path(j), os.O_RDONLY, 0)
	return
}

func (fs fileStorage) OpenWriter(j *Job) (wc io.WriteCloser, err os.Error) {
	wc, err = os.Open(fs.path(j), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	return
}

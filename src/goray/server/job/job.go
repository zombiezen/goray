//
//  goray/server/job/job.go
//  goray
//
//  Created by Ross Light on 2011-03-14.
//

// The job package specifies a consistent description of a render job.
package job

import (
	"image/png"
	"io"
	"os"
	"sync"
	"goray/core/scene"
	"goray/core/integrator"
	"goray/std/yamlscene"
)

type Job struct {
	Name       string
	YAML       io.Reader
	OutputFile io.WriteCloser

	status StatusCode
	err    os.Error
	lock   sync.RWMutex
	cond   *sync.Cond
}

func (job *Job) Status() StatusCode {
	job.lock.RLock()
	defer job.lock.RUnlock()
	return job.status
}

func (job *Job) Error() os.Error {
	job.lock.RLock()
	defer job.lock.RUnlock()
	return job.err
}

func (job *Job) StatusChan() <-chan StatusCode {
	ch := make(chan StatusCode)
	go func() {
		defer close(ch)
		job.cond.L.Lock()
		defer job.cond.L.Unlock()

		stat := job.status
		for !stat.Finished() {
			job.cond.Wait()
			if job.status != stat {
				ch <- job.status
			}
			stat = job.status
		}
		ch <- stat
	}()
	return ch
}

func (job *Job) changeStatus(stat StatusCode) {
	job.lock.Lock()
	defer job.lock.Unlock()
	job.status = stat
	job.cond.Broadcast()
}

func (job *Job) Render() (err os.Error) {
	defer job.OutputFile.Close()
	job.changeStatus(StatusRendering)
	defer func() {
		if err != nil {
			job.changeStatus(StatusError)
		} else {
			job.changeStatus(StatusDone)
		}
	}()

	sc := scene.New()
	integ, err := yamlscene.Load(job.YAML, sc)
	if err != nil {
		return
	}
	sc.Update()
	outputImage := integrator.Render(sc, integ, nil)
	err = png.Encode(job.OutputFile, outputImage)
	return
}

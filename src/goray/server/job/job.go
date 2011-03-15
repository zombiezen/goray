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
	"goray/core/render"
	"goray/core/scene"
	"goray/core/integrator"
	"goray/std/yamlscene"
	"goray/time"
)

type Job struct {
	Name       string
	YAML       io.Reader
	OutputFile io.WriteCloser

	status Status
	lock   sync.RWMutex
	cond   *sync.Cond
}

func (job *Job) Status() Status {
	job.lock.RLock()
	defer job.lock.RUnlock()
	return job.status
}

func (job *Job) StatusChan() <-chan Status {
	ch := make(chan Status)
	go func() {
		defer close(ch)
		job.cond.L.Lock()
		defer job.cond.L.Unlock()

		stat := job.status
		ch <- stat
		for !stat.Finished() {
			job.cond.Wait()
			if job.status.Code != stat.Code {
				ch <- job.status
			}
			stat = job.status
		}
	}()
	return ch
}

func (job *Job) changeStatus(stat Status) {
	job.lock.Lock()
	defer job.lock.Unlock()
	job.status = stat
	job.cond.Broadcast()
}

func (job *Job) Render() (err os.Error) {
	defer job.OutputFile.Close()

	var status Status
	status.Code = StatusRendering
	job.changeStatus(status)
	defer func() {
		if err != nil {
			status.Code, status.Error = StatusError, err
			job.changeStatus(status)
		} else {
			status.Code = StatusDone
			job.changeStatus(status)
		}
	}()

	sc := scene.New()
	var integ integrator.Integrator
	status.ReadTime = time.Stopwatch(func() {
		integ, err = yamlscene.Load(job.YAML, sc)
	})
	if err != nil {
		return
	}
	status.UpdateTime = time.Stopwatch(func() {
		sc.Update()
	})
	var outputImage *render.Image
	status.RenderTime = time.Stopwatch(func() {
		outputImage = integrator.Render(sc, integ, nil)
	})
	status.WriteTime = time.Stopwatch(func() {
		err = png.Encode(job.OutputFile, outputImage)
	})
	return
}

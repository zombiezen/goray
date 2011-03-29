//
//  goray/job/job.go
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
	Name   string
	Source io.Reader

	status Status
	lock   sync.RWMutex
	cond   *sync.Cond
}

// New returns a newly allocated job structure.
func New(name string, source io.Reader) (job *Job) {
	job = &Job{
		Name:   name,
		Source: source,
	}
	job.cond = sync.NewCond(job.lock.RLocker())
	return
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

func (job *Job) ChangeStatus(stat Status) {
	job.lock.Lock()
	defer job.lock.Unlock()
	job.status = stat
	job.cond.Broadcast()
}

func (job *Job) Render(w io.Writer) (err os.Error) {
	var status Status
	defer func() {
		if err != nil {
			status.Code, status.Error = StatusError, err
			job.ChangeStatus(status)
		} else {
			status.Code = StatusDone
			job.ChangeStatus(status)
		}
	}()

	// 1. Read
	status.Code = StatusReading
	job.ChangeStatus(status)
	sc := scene.New()
	var integ integrator.Integrator
	status.ReadTime = time.Stopwatch(func() {
		integ, err = yamlscene.Load(job.Source, sc)
	})
	if err != nil {
		return
	}
	// 2. Update
	status.Code = StatusUpdating
	job.ChangeStatus(status)
	status.UpdateTime = time.Stopwatch(func() {
		sc.Update()
	})
	// 3. Render
	var outputImage *render.Image
	status.Code = StatusRendering
	job.ChangeStatus(status)
	status.RenderTime = time.Stopwatch(func() {
		outputImage = integrator.Render(sc, integ, nil)
	})
	// 4. Write
	status.Code = StatusWriting
	job.ChangeStatus(status)
	status.WriteTime = time.Stopwatch(func() {
		err = png.Encode(w, outputImage)
	})
	return
}

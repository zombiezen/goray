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

// Package job specifies a consistent description of a render job.
package job

import (
	"io"
	"sync"
	"time"

	"zombiezen.com/go/goray/internal/goray"
	"zombiezen.com/go/goray/internal/intersect"
	"zombiezen.com/go/goray/internal/log"
	"zombiezen.com/go/goray/internal/yamlscene"
)

type Job struct {
	Name      string
	Source    io.Reader
	Params    yamlscene.Params
	SceneLog  log.Logger
	RenderLog log.Logger

	status Status
	lock   sync.RWMutex
	cond   *sync.Cond
}

// New returns a newly allocated job structure.
func New(name string, source io.Reader, params yamlscene.Params) (job *Job) {
	job = &Job{
		Name:   name,
		Source: source,
		Params: params,
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

func (job *Job) Render(w io.Writer) (err error) {
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
	sc := goray.NewScene(goray.IntersecterBuilder(intersect.NewKD), job.SceneLog)
	var integ goray.Integrator
	status.ReadTime = stopwatch(func() {
		integ, err = yamlscene.Load(job.Source, sc, job.Params)
	})
	if err != nil {
		return
	}

	// 2. Update
	status.Code = StatusUpdating
	job.ChangeStatus(status)
	status.UpdateTime = stopwatch(func() {
		sc.Update()
	})

	// 3. Render
	var outputImage *goray.Image
	status.Code = StatusRendering
	job.ChangeStatus(status)
	status.RenderTime = stopwatch(func() {
		outputImage = goray.Render(sc, integ, job.RenderLog)
	})

	// 4. Write
	status.Code = StatusWriting
	job.ChangeStatus(status)
	format, ok := job.Params["OutputFormat"].(Format)
	if !ok {
		format = FormatMap[DefaultFormat]
	}
	status.WriteTime = stopwatch(func() {
		err = format.Encode(w, outputImage)
	})
	return
}

// stopwatch calls a function and returns how long it took for the function to return.
func stopwatch(f func()) time.Duration {
	startTime := time.Now()
	f()
	endTime := time.Now()
	return endTime.Sub(startTime)
}

//
//  goray/server/jobs.go
//  goray
//
//  Created by Ross Light on 2011-02-05.
//

package server

import (
	"fmt"
	"image/png"
	"io"
	"os"
	pathutil "path"
	"sync"
	"goray/core/scene"
	"goray/core/integrator"
	"goray/std/yamlscene"
)

type Job struct {
	Name       string
	YAML       io.Reader
	OutputFile io.WriteCloser
	Done       bool
	Status     chan bool
}

func (job Job) Render() (err os.Error) {
	defer job.OutputFile.Close()
	defer func() {
		job.Done = true
		job.Status <- job.Done
		close(job.Status)
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

type JobManager struct {
	OutputDirectory string
	jobs            map[string]Job
	nextNum         int
	lock            sync.RWMutex
}

func (manager *JobManager) New(yaml io.Reader) (j Job, err os.Error) {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	// Get next name
	name := fmt.Sprintf("%04d", manager.nextNum)
	manager.nextNum++
	// Open output file
	f, err := os.Open(
		pathutil.Join(manager.OutputDirectory, name+".png"),
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666,
	)
	if err != nil {
		return
	}
	// Create job
	j = Job{
		Name:       name,
		YAML:       yaml,
		OutputFile: f,
		Status:     make(chan bool),
	}
	if manager.jobs == nil {
		manager.jobs = make(map[string]Job)
	}
	manager.jobs[name] = j
	return
}

func (manager *JobManager) Get(name string) (j Job, ok bool) {
	manager.lock.RLock()
	defer manager.lock.RUnlock()

	if manager.jobs == nil {
		return
	}
	j, ok = manager.jobs[name]
	return
}

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
	"path/filepath"
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
	Cond       *sync.Cond
	statusLock sync.RWMutex
}

func (job *Job) Render() (err os.Error) {
	defer job.OutputFile.Close()
	defer func() {
		job.statusLock.Lock()
		job.Done = true
		job.statusLock.Unlock()
		job.Cond.Broadcast()
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

// JobManager maintains a render job queue and records completed jobs.
type JobManager struct {
	OutputDirectory string

	jobs     map[string]*Job
	jobQueue chan *Job
	nextNum  int
	lock     sync.RWMutex
}

// NewJobManager creates a new, initialized job manager.
func NewJobManager(outdir string, queueSize int) (manager *JobManager) {
	manager = &JobManager{OutputDirectory: outdir}
	manager.Init(queueSize)
	return
}

// Init initializes the manager.  This function is called automatically by
// NewJobManager.
func (manager *JobManager) Init(queueSize int) {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	if manager.jobQueue != nil {
		close(manager.jobQueue)
	}
	manager.jobs = make(map[string]*Job)
	manager.jobQueue = make(chan *Job, queueSize)
	manager.nextNum = 0
}

// New creates a new job and adds it to the job queue.
func (manager *JobManager) New(yaml io.Reader) (j *Job, err os.Error) {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	// Get next name
	name := fmt.Sprintf("%04d", manager.nextNum)
	// Open output file
	f := &deferredFile{
		Path: filepath.Join(manager.OutputDirectory, name+".png"),
		Flag: os.O_WRONLY | os.O_CREATE | os.O_TRUNC,
		Perm: 0666,
	}
	// Create job
	j = &Job{
		Name:       name,
		YAML:       yaml,
		OutputFile: f,
	}
	j.Cond = sync.NewCond(j.statusLock.RLocker())
	// Try to add job to queue
	select {
	case manager.jobQueue <- j:
		manager.jobs[name] = j
		manager.nextNum++
	default:
		err = os.NewError("Job queue is full")
	}
	return
}

// Get returns a job with the given name.
func (manager *JobManager) Get(name string) (j *Job, ok bool) {
	manager.lock.RLock()
	defer manager.lock.RUnlock()

	if manager.jobs == nil {
		return
	}
	j, ok = manager.jobs[name]
	return
}

// Stop causes the manager to stop accepting new jobs.
func (manager *JobManager) Stop() {
	close(manager.jobQueue)
}

// RenderJobs renders jobs in the queue until Stop is called.
func (manager *JobManager) RenderJobs() {
	for job := range manager.jobQueue {
		job.Render()
	}
}

// A deferredFile holds the parameters to an open call, and only opens the file
// when a write occurs.
type deferredFile struct {
	Path string
	Flag int
	Perm uint32
	file *os.File
}

// open opens the underlying file and returns any error.  This does nothing if
// the file was already opened.
func (f *deferredFile) open() (err os.Error) {
	if f.file != nil {
		return
	}
	f.file, err = os.Open(f.Path, f.Flag, f.Perm)
	return
}

func (f *deferredFile) Write(p []byte) (n int, err os.Error) {
	if f.file == nil {
		err = f.open()
		if err != nil {
			return
		}
	}
	return f.file.Write(p)
}

func (f *deferredFile) Close() (err os.Error) {
	if f.file == nil {
		return
	}
	return f.file.Close()
}

//
//  goray/server/job/manager.go
//  goray
//
//  Created by Ross Light on 2011-03-14.
//

package job

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// Manager maintains a render job queue and records completed jobs.
type Manager struct {
	OutputDirectory string

	jobs     map[string]*Job
	jobQueue chan *Job
	nextNum  int
	lock     sync.RWMutex
}

// NewManager creates a new, initialized job manager.
func NewManager(outdir string, queueSize int) (manager *Manager) {
	manager = &Manager{OutputDirectory: outdir}
	manager.Init(queueSize)
	return
}

// Init initializes the manager.  This function is called automatically by
// NewJobManager.
func (manager *Manager) Init(queueSize int) {
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
func (manager *Manager) New(yaml io.Reader) (j *Job, err os.Error) {
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
	j.cond = sync.NewCond(j.lock.RLocker())
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
func (manager *Manager) Get(name string) (j *Job, ok bool) {
	manager.lock.RLock()
	defer manager.lock.RUnlock()
	j, ok = manager.jobs[name]
	return
}

func (manager *Manager) List() (jobs []*Job) {
	manager.lock.RLock()
	defer manager.lock.RUnlock()
	jobs = make([]*Job, 0, len(manager.jobs))
	for _, j := range manager.jobs {
		jobs = append(jobs, j)
	}
	return
}

// Stop causes the manager to stop accepting new jobs.
func (manager *Manager) Stop() {
	close(manager.jobQueue)
}

// RenderJobs renders jobs in the queue until Stop is called.
func (manager *Manager) RenderJobs() {
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

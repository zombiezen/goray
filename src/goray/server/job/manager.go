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
	"sync"
)

// Manager maintains a render job queue and records completed jobs.
type Manager struct {
	Storage  Storage
	jobs     map[string]*Job
	jobQueue chan *Job
	nextNum  int
	lock     sync.RWMutex
}

// NewManager creates a new, initialized job manager.
func NewManager(storage Storage, queueSize int) (manager *Manager) {
	manager = &Manager{Storage: storage}
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
	name := fmt.Sprintf("%04d", manager.nextNum)
	j = New(name, yaml)
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
		w, err := manager.Storage.OpenWriter(job)
		if err == nil {
			job.Render(w)
			w.Close()
		} else {
			job.ChangeStatus(Status{Code: StatusError, Error: err})
		}
	}
}

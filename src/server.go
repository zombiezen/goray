//
//  server.go
//  goray
//
//  Created by Ross Light on 2011-02-05.
//

package main

import (
	"bytes"
	"flag"
	"fmt"
	"http"
	"io"
	"image/png"
	"os"
	pathutil "path"
	"strings"
	"sync"
	"template"
)

import (
	"goray/core/scene"
	"goray/core/integrator"
	"goray/std/yamlscene"
)

var fmap = template.FormatterMap{
	"":     template.HTMLFormatter,
	"str":  template.StringFormatter,
	"html": template.HTMLFormatter,
}

const submitTemplateSource = `<!doctype html>
<html>
	<head>
		<title>Goray</title>
	</head>
	<body>
		<h1>Goray Job Submission</h1>
		<p>Submit a render document below!</p>
		<form action="." method="POST">
			<p><textarea name="data" rows="30" cols="80"></textarea></p>
			<p><input type="submit"></p>
		</form>
	</body>
</html>`

const jobTemplateSource = `<!doctype html>
<html>
	<head>
		<title>Goray - Job {Name}</title>
	</head>
	<body>
		<h1>Job #{Name}</h1>
		<p>Wait a little bit...</p>
		<p><a href="/output/{Name}.png">And then click this.</a></p>
	</body>
</html>`

var submitTemplate = template.MustParse(submitTemplateSource, fmap)
var jobTemplate = template.MustParse(jobTemplateSource, fmap)

type Job struct {
	Name       string
	YAML       io.Reader
	OutputFile io.WriteCloser
}

func (job Job) Render() (err os.Error) {
	defer job.OutputFile.Close()
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

func SubmitJob(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}

	switch req.Method {
	case "GET":
		w.SetHeader("Content-Type", "text/html; charset=utf-8")
		submitTemplate.Execute(nil, w)
	case "POST":
		j, _ := manager.New(bytes.NewBufferString(req.FormValue("data")))
		go j.Render()
		http.Redirect(w, req, "/job/"+j.Name, http.StatusMovedPermanently)
	default:
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
	}
}

func ViewJob(w http.ResponseWriter, req *http.Request) {
	jobName := req.URL.Path[strings.LastIndex(req.URL.Path, "/")+1:]
	job, ok := manager.Get(jobName)
	if ok {
		w.SetHeader("Content-Type", "text/html; charset=utf-8")
		jobTemplate.Execute(job, w)
	} else {
		http.NotFound(w, req)
	}
}

var manager *JobManager

func main() {
	flag.Parse()

	if flag.NArg() != 2 {
		return
	}

	manager = &JobManager{
		OutputDirectory: flag.Arg(1),
	}

	http.HandleFunc("/", SubmitJob)
	http.HandleFunc("/job/", ViewJob)
	http.Handle("/output/", http.FileServer(manager.OutputDirectory, "/output/"))
	http.ListenAndServe(flag.Arg(0), nil)
}

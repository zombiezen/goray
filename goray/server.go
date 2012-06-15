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

package main

import (
	"bytes"
	"flag"
	"html/template"
	"io"
	"net/http"
	"net/textproto"
	"path/filepath"
	"strconv"

	"code.google.com/p/go.net/websocket"
	"code.google.com/p/gorilla/mux"

	"bitbucket.org/zombiezen/goray/job"
	"bitbucket.org/zombiezen/goray/logging"
	"bitbucket.org/zombiezen/goray/std/textures/image/fileloader"
	"bitbucket.org/zombiezen/goray/std/yamlscene"
)

var (
	router      *mux.Router
	jobManager  *job.Manager
	baseParams  yamlscene.Params
	templates   *template.Template
	logRecorder *logging.CircularHandler
)

const defaultOutputRoot = "output"

const (
	dataTemplateSubdir = "templates"
	dataStaticSubdir   = "static"
)

func httpServer() int {
	// Parse arguments
	if flag.NArg() != 0 {
		printInstructions()
		return 1
	}
	if outputPath == "" {
		outputPath = defaultOutputRoot
	}

	// Init job manager
	storage, err := job.NewFileStorage(outputPath)
	if err != nil {
		logging.MainLog.Critical("FileStorage: %v", err)
		return 1
	}
	jobManager = job.NewManager(storage, 5)

	// Parse templates
	templates, err = template.ParseGlob(filepath.Join(dataRoot, dataTemplateSubdir, "*.html"))
	if err != nil {
		logging.MainLog.Critical("Server: %v", err)
		return 1
	}

	// Routes
	router = mux.NewRouter()
	router.HandleFunc("/", handleIndex).Name("index")
	router.HandleFunc("/license", handleLicense).Name("license")
	router.HandleFunc("/job/{job:[0-9]+}", handleViewJob).Name("view")
	router.HandleFunc("/submit", handleSubmitJob).Name("submit")
	router.HandleFunc("/log", handleLog).Name("log")
	router.Handle("/status", websocket.Handler(func(ws *websocket.Conn) { handleStatus(ws) })).Name("status")
	router.HandleFunc("/output/{job:[0-9]+}", handleOutput).Name("output")
	fs := http.FileServer(http.Dir(filepath.Join(dataRoot, dataStaticSubdir)))
	router.HandleFunc("/static/{path:.*}", func(w http.ResponseWriter, req *http.Request) {
		req.URL.Path = mux.Vars(req)["path"]
		fs.ServeHTTP(w, req)
	}).Name("static")

	// Init logging
	logRecorder = logging.NewCircularHandler(100)
	logging.MainLog.AddHandler(logRecorder)

	// Start up job rendering
	baseParams = yamlscene.Params{
		"ImageLoader": fileloader.New(imagePath),
	}
	go jobManager.RenderJobs()

	// Run HTTP server
	logging.MainLog.Info("Starting HTTP server")
	err = http.ListenAndServe(httpAddress, router)
	if err != nil {
		logging.MainLog.Critical("ListenAndServe: %v", err)
		return 1
	}
	return 0
}

func handleIndex(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.ExecuteTemplate(w, "index.html", jobManager.List())
}

func handleLicense(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.ExecuteTemplate(w, "license.html", nil)
}

func handleSubmitJob(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		templates.ExecuteTemplate(w, "submit.html", nil)
	case "POST":
		j, err := jobManager.New(bytes.NewBufferString(req.FormValue("data")), baseParams)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logging.MainLog.Info("Created job %s", j.Name)
		http.Redirect(w, req, "/job/"+j.Name, http.StatusMovedPermanently)
	default:
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
	}
}

func handleLog(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.ExecuteTemplate(w, "log.html", logRecorder.Records())
}

func handleViewJob(w http.ResponseWriter, req *http.Request) {
	j, ok := jobManager.Get(mux.Vars(req)["job"])
	if ok {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Check to see whether the job is done
		status := j.Status()

		// Render appropriate template
		switch status.Code {
		case job.StatusDone:
			templates.ExecuteTemplate(w, "job.html", j)
		case job.StatusError:
			templates.ExecuteTemplate(w, "job-error.html", j)
		default:
			templates.ExecuteTemplate(w, "job-waiting.html", j)
		}
	} else {
		http.NotFound(w, req)
	}
}

func handleStatus(ws *websocket.Conn) {
	defer ws.Close()
	conn := textproto.NewConn(ws)
	jobName, err := conn.ReadLine()
	if err != nil {
		return
	}

	// Find job
	j, found := jobManager.Get(jobName)
	if !found {
		conn.PrintfLine("404 Job not found")
		return
	}

	// Notify when job is finished
	for status := range j.StatusChan() {
		conn.PrintfLine("%d %s", int(status.Code), status.Code)
		switch status.Code {
		case job.StatusDone:
			conn.PrintfLine("")
			conn.PrintfLine("%v", status.TotalTime())
		case job.StatusError:
			conn.PrintfLine("")
			conn.PrintfLine("%s", status.Error)
		}
	}
}

func handleOutput(w http.ResponseWriter, req *http.Request) {
	j, ok := jobManager.Get(mux.Vars(req)["job"])
	if ok && j.Status().Code == job.StatusDone {
		w.Header().Set("Content-Type", "image/png; charset=utf-8")
		r, err := jobManager.Storage.OpenReader(j)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if seeker, ok := r.(io.Seeker); ok {
			size, err := seeker.Seek(0, 2)
			if err == nil {
				w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
			}
			seeker.Seek(0, 0)
		}
		io.Copy(w, r)
	} else {
		http.NotFound(w, req)
	}
}

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

// Package server provides an HTTP front-end for goray.
package server

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"net/textproto"
	"path/filepath"
	"strconv"

	"bitbucket.org/zombiezen/goray/job"
	"bitbucket.org/zombiezen/goray/logging"
	"bitbucket.org/zombiezen/goray/server/urls"
	"bitbucket.org/zombiezen/goray/std/yamlscene"

	"code.google.com/p/go.net/websocket"
)

type Server struct {
	Resolver   *urls.RegexResolver
	DataRoot   string
	JobManager *job.Manager
	BaseParams yamlscene.Params

	templates   *template.Template
	logRecorder *logging.CircularHandler
}

func New(manager *job.Manager, data string) (*Server, error) {
	t, err := template.ParseGlob(filepath.Join(data, "templates", "*.html"))
	if err != nil {
		return nil, err
	}
	s := &Server{
		DataRoot:    data,
		JobManager:  manager,
		templates:   t,
		logRecorder: logging.NewCircularHandler(100),
	}
	s.Resolver = urls.Patterns(``,
		urls.New(`^$`, serverView{s, (*Server).handleIndex}, "index"),
		urls.New(`^license$`, serverView{s, (*Server).handleLicense}, "license"),
		urls.New(`^job/([0-9]+)$`, serverView{s, (*Server).handleViewJob}, "view"),
		urls.New(`^submit$`, serverView{s, (*Server).handleSubmitJob}, "submit"),
		urls.New(`^log$`, serverView{s, (*Server).handleLog}, "log"),
		urls.New(`^status$`, urls.HandlerView{websocket.Handler(func(ws *websocket.Conn) { s.handleStatus(ws) })}, "status"),
		urls.New(`^static/`, urls.HandlerView{http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(data, "static"))))}, "static"),
		urls.New(`^output/([0-9]+)$`, serverView{s, (*Server).handleOutput}, "output"),
	)
	logging.MainLog.AddHandler(s.logRecorder)
	go s.JobManager.RenderJobs()
	return s, nil
}

func (server *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	server.Resolver.ServeHTTP(w, req)
}

func (server *Server) handleIndex(w http.ResponseWriter, req *http.Request, args []string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	server.templates.ExecuteTemplate(w, "index.html", server.JobManager.List())
}

func (server *Server) handleLicense(w http.ResponseWriter, req *http.Request, args []string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	server.templates.ExecuteTemplate(w, "license.html", nil)
}

func (server *Server) handleSubmitJob(w http.ResponseWriter, req *http.Request, args []string) {
	switch req.Method {
	case "GET":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		server.templates.ExecuteTemplate(w, "submit.html", nil)
	case "POST":
		j, err := server.JobManager.New(bytes.NewBufferString(req.FormValue("data")), server.BaseParams)
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

func (server *Server) handleLog(w http.ResponseWriter, req *http.Request, args []string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	server.templates.ExecuteTemplate(w, "log.html", server.logRecorder.Records())
}

func (server *Server) handleViewJob(w http.ResponseWriter, req *http.Request, args []string) {
	j, ok := server.JobManager.Get(args[0])
	if ok {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Check to see whether the job is done
		status := j.Status()

		// Render appropriate template
		switch status.Code {
		case job.StatusDone:
			server.templates.ExecuteTemplate(w, "job.html", j)
		case job.StatusError:
			server.templates.ExecuteTemplate(w, "job-error.html", j)
		default:
			server.templates.ExecuteTemplate(w, "job-waiting.html", j)
		}
	} else {
		http.NotFound(w, req)
	}
}

func (server *Server) handleStatus(ws *websocket.Conn) {
	defer ws.Close()
	conn := textproto.NewConn(ws)
	jobName, err := conn.ReadLine()
	if err != nil {
		return
	}

	// Find job
	j, found := server.JobManager.Get(jobName)
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

func (server *Server) handleOutput(w http.ResponseWriter, req *http.Request, args []string) {
	j, ok := server.JobManager.Get(args[0])
	if ok && j.Status().Code == job.StatusDone {
		w.Header().Set("Content-Type", "image/png; charset=utf-8")
		r, err := server.JobManager.Storage.OpenReader(j)
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

type serverView struct {
	Server *Server
	Func   func(*Server, http.ResponseWriter, *http.Request, []string)
}

func (v serverView) Render(w http.ResponseWriter, req *http.Request, args []string) {
	v.Func(v.Server, w, req, args)
}

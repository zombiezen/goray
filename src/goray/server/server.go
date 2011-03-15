//
//  goray/server/server.go
//  goray
//
//  Created by Ross Light on 2011-02-05.
//

/*
	The server package provides an HTTP front-end for goray.
*/
package server

import (
	"bytes"
	"http"
	"io"
	"net/textproto"
	"path/filepath"
	"strings"
	"strconv"
	"websocket"

	"goray/server/job"
)

type Server struct {
	*http.ServeMux

	DataRoot   string
	JobManager *job.Manager
	templates  *TemplateLoader
}

func New(manager *job.Manager, data string) (s *Server) {
	s = &Server{
		ServeMux:   http.NewServeMux(),
		DataRoot:   data,
		JobManager: manager,
		templates:  &TemplateLoader{Root: filepath.Join(data, "templates")},
	}
	s.Handle("/", serverHandler{s, (*Server).handleIndex})
	s.Handle("/job/", serverHandler{s, (*Server).handleViewJob})
	s.Handle("/submit", serverHandler{s, (*Server).handleSubmitJob})
	s.Handle("/status", websocket.Handler(func(ws *websocket.Conn) {
		s.handleStatus(ws)
	}))
	s.Handle("/static/", http.FileServer(filepath.Join(data, "static"), "/static/"))
	s.Handle("/output/", serverHandler{s, (*Server).handleOutput})
	go s.JobManager.RenderJobs()
	return
}

func (server *Server) handleIndex(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	w.SetHeader("Content-Type", "text/html; charset=utf-8")
	server.templates.RenderResponse(w, "index.html", map[string]interface{}{
		"Jobs": server.JobManager.List(),
	})
}

func (server *Server) handleSubmitJob(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		w.SetHeader("Content-Type", "text/html; charset=utf-8")
		server.templates.RenderResponse(w, "submit.html", nil)
	case "POST":
		j, err := server.JobManager.New(bytes.NewBufferString(req.FormValue("data")))
		if err != nil {
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, req, "/job/"+j.Name, http.StatusMovedPermanently)
	default:
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
	}
}

func (server *Server) handleViewJob(w http.ResponseWriter, req *http.Request) {
	jobName := req.URL.Path[strings.LastIndex(req.URL.Path, "/")+1:]
	j, ok := server.JobManager.Get(jobName)
	if ok {
		w.SetHeader("Content-Type", "text/html; charset=utf-8")
		// Check to see whether the job is done
		status := j.Status()
		// Render appropriate template
		switch status.Code {
		case job.StatusDone:
			server.templates.RenderResponse(w, "job.html", j)
		case job.StatusError:
			server.templates.RenderResponse(w, "job-error.html", j)
		default:
			server.templates.RenderResponse(w, "job-waiting.html", j)
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

func (server *Server) handleOutput(w http.ResponseWriter, req *http.Request) {
	jobName := req.URL.Path[strings.LastIndex(req.URL.Path, "/")+1:]
	j, ok := server.JobManager.Get(jobName)
	if ok && j.Status().Code == job.StatusDone {
		w.SetHeader("Content-Type", "image/png; charset=utf-8")
		r, err := server.JobManager.Storage.OpenReader(j)
		if err != nil {
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}
		if seeker, ok := r.(io.Seeker); ok {
			size, err := seeker.Seek(0, 2)
			if err == nil {
				w.SetHeader("Content-Length", strconv.Itoa64(size))
			}
			seeker.Seek(0, 0)
		}
		io.Copy(w, r)
	} else {
		http.NotFound(w, req)
	}
}

type serverHandler struct {
	Server *Server
	Func   func(*Server, http.ResponseWriter, *http.Request)
}

func (h serverHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.Func(h.Server, w, req)
}

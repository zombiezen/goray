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
	"net/textproto"
	pathutil "path"
	"strings"
	"websocket"
)

type Server struct {
	*http.ServeMux

	DataRoot   string
	JobManager *JobManager
	templates  *TemplateLoader
}

func New(output, data string) (s *Server) {
	s = &Server{
		ServeMux:   http.NewServeMux(),
		DataRoot:   data,
		JobManager: NewJobManager(output, 5),
		templates:  &TemplateLoader{Root: pathutil.Join(data, "templates")},
	}
	s.Handle("/", serverHandler{s, (*Server).handleSubmitJob})
	s.Handle("/job/", serverHandler{s, (*Server).handleViewJob})
	s.Handle("/status", websocket.Handler(func(ws *websocket.Conn) {
		s.handleStatus(ws)
	}))
	s.Handle("/static/", http.FileServer(pathutil.Join(data, "static"), "/static/"))
	s.Handle("/output/", http.FileServer(output, "/output/"))
	go s.JobManager.RenderJobs()
	return
}

func (server *Server) handleSubmitJob(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}

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
	job, ok := server.JobManager.Get(jobName)
	if ok {
		w.SetHeader("Content-Type", "text/html; charset=utf-8")
		server.templates.RenderResponse(w, "job.html", job)
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
	job, found := server.JobManager.Get(jobName)
	if !found {
		conn.PrintfLine("404 Job not found")
		return
	}
	conn.PrintfLine("200 Job found")
	// Notify when job is finished
	job.Cond.L.Lock()
	for !job.Done {
		job.Cond.Wait()
	}
	conn.PrintfLine("done")
	job.Cond.L.Unlock()
}

type serverHandler struct {
	Server *Server
	Func   func(*Server, http.ResponseWriter, *http.Request)
}

func (h serverHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.Func(h.Server, w, req)
}

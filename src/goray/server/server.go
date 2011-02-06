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
		JobManager: &JobManager{OutputDirectory: output},
		templates:  &TemplateLoader{Root: pathutil.Join(data, "templates")},
	}
	s.Handle("/", serverHandler{s, (*Server).handleSubmitJob})
	s.Handle("/job/", serverHandler{s, (*Server).handleViewJob})
	s.Handle("/status", websocket.Handler(func(ws *websocket.Conn) {
		s.handleStatus(ws)
	}))
	s.Handle("/static/", http.FileServer(pathutil.Join(data, "static"), "/static/"))
	s.Handle("/output/", http.FileServer(output, "/output/"))
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
		server.templates.RenderResponse("submit.html", nil, w)
	case "POST":
		j, _ := server.JobManager.New(bytes.NewBufferString(req.FormValue("data")))
		go j.Render()
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
		server.templates.RenderResponse("job.html", job, w)
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
	job, found := server.JobManager.Get(jobName)
	if !found {
		conn.PrintfLine("404 Job not found")
		return
	}

	conn.PrintfLine("200 Job found")
	if job.Done {
		conn.PrintfLine("done")
		return
	}

	// Wait for job to finish
	<-job.Status
	conn.PrintfLine("done")
}

type serverHandler struct {
	Server *Server
	Func   func(*Server, http.ResponseWriter, *http.Request)
}

func (h serverHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.Func(h.Server, w, req)
}

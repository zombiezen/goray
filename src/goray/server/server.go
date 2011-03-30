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
	"strconv"
	"strings"
	"websocket"

	"goray/job"
	"goray/server/urls"
)

type Server struct {
	Resolver   *urls.RegexResolver
	DataRoot   string
	JobManager *job.Manager
	templates  *TemplateLoader
	blocks map[string]string
}

func New(manager *job.Manager, data string) (s *Server) {
	s = &Server{
		DataRoot:   data,
		JobManager: manager,
		templates:  &TemplateLoader{Root: filepath.Join(data, "templates")},
	}
	s.Resolver = urls.Patterns(``,
		urls.New(`^$`, serverView{s, (*Server).handleIndex}, "index"),
		urls.New(`^job/([0-9]+)$`, serverView{s, (*Server).handleViewJob}, "view"),
		urls.New(`^submit$`, serverView{s, (*Server).handleSubmitJob}, "submit"),
		urls.New(`^status$`, urls.HandlerView{websocket.Handler(func(ws *websocket.Conn) { s.handleStatus(ws) })}, "status"),
		urls.New(`^static/`, urls.HandlerView{http.FileServer(filepath.Join(data, "static"), "/static/")}, "static"),
		urls.New(`^output/([0-9]+)$`, serverView{s, (*Server).handleOutput}, "output"),
	)
	s.blocks = map[string]string{
		"Head": "blocks/head.html",
		"Header": "blocks/header.html",
		"Footer": "blocks/footer.html",
	}
	for k, path := range s.blocks {
		b := new(bytes.Buffer)
		s.templates.Render(b, path, nil)
		s.blocks[k] = strings.TrimSpace(b.String())
	}
	go s.JobManager.RenderJobs()
	return
}

func (server *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	server.Resolver.ServeHTTP(w, req)
}

func (server *Server) handleIndex(w http.ResponseWriter, req *http.Request, args []string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	server.templates.RenderResponse(w, "index.html", map[string]interface{}{
		"Blocks": server.blocks,
		"Jobs": server.JobManager.List(),
	})
}

func (server *Server) handleSubmitJob(w http.ResponseWriter, req *http.Request, args []string) {
	switch req.Method {
	case "GET":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		server.templates.RenderResponse(w, "submit.html", map[string]interface{}{
			"Blocks": server.blocks,
		})
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

func (server *Server) handleViewJob(w http.ResponseWriter, req *http.Request, args []string) {
	j, ok := server.JobManager.Get(args[0])
	if ok {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Check to see whether the job is done
		status := j.Status()
		// Render appropriate template
		switch status.Code {
		case job.StatusDone:
			server.templates.RenderResponse(w, "job.html", map[string]interface{}{
				"Blocks": server.blocks,
				"Job": j,
			})
		case job.StatusError:
			server.templates.RenderResponse(w, "job-error.html", map[string]interface{}{
				"Blocks": server.blocks,
				"Job": j,
			})
		default:
			server.templates.RenderResponse(w, "job-waiting.html", map[string]interface{}{
				"Blocks": server.blocks,
				"Job": j,
			})
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
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}
		if seeker, ok := r.(io.Seeker); ok {
			size, err := seeker.Seek(0, 2)
			if err == nil {
				w.Header().Set("Content-Length", strconv.Itoa64(size))
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

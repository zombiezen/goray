//
//  goray/server/templates.go
//  goray
//
//  Created by Ross Light on 2011-02-05.
//

package server

import (
	"bytes"
	"http"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"template"
)

var fmap = template.FormatterMap{
	"":     template.HTMLFormatter,
	"str":  template.StringFormatter,
	"html": template.HTMLFormatter,
}

type TemplateLoader struct {
	Root  string
	cache map[string]*template.Template
	lock  sync.RWMutex
}

func (loader *TemplateLoader) Get(name string) (templ *template.Template, err os.Error) {
	// Check cache first
	var cacheHit bool
	loader.lock.RLock()
	if loader.cache != nil {
		templ, cacheHit = loader.cache[name]
	}
	loader.lock.RUnlock()
	if cacheHit {
		return
	}
	// Load template
	path := filepath.Join(loader.Root, filepath.FromSlash(name))
	templ = template.New(fmap)
	templ.SetDelims("{{", "}}")
	err = templ.ParseFile(path)
	if err != nil {
		return nil, err
	}
	// Save template to cache
	// Yes, another thread may have already read in the template. However, the
	// end result is really the same (and yes, there could be a disk write in
	// between all of this).
	loader.lock.Lock()
	if loader.cache == nil {
		loader.cache = make(map[string]*template.Template)
	}
	loader.cache[name] = templ
	loader.lock.Unlock()
	return
}

func (loader *TemplateLoader) Render(w io.Writer, name string, data interface{}) (err os.Error) {
	t, err := loader.Get(name)
	if err != nil {
		return
	}
	return t.Execute(w, data)
}

func (loader *TemplateLoader) RenderResponse(rw http.ResponseWriter, name string, data interface{}) (err os.Error) {
	// Buffer template output
	buf := new(bytes.Buffer)
	err = loader.Render(buf, name, data)
	if err != nil {
		return
	}
	// Write to response
	rw.SetHeader("Content-Length", strconv.Itoa(buf.Len()))
	_, err = io.Copy(rw, buf)
	return
}

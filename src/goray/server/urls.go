//
//  goray/server/urls.go
//  goray
//
//  Created by Ross Light on 2011-03-15.
//

// The urls package provides a regular expression-based URL router.
package urls

import (
	"http"
	"path"
	"regexp"
)

// A View is an HTTP handler that accepts string arguments from a URL.
type View interface {
	Render(http.ResponseWriter, *http.Request, []string)
}

// A ViewFunc is a function-based view.
type ViewFunc func(http.ResponseWriter, *http.Request, []string)

func (f ViewFunc) Render(w http.ResponseWriter, req *http.Request, args []string) {
	f(w, req, args)
}

// A HandlerView wraps an http.Handler so that it can be used inside the resolver context.
type HandlerView struct {
	http.Handler
}

func (v HandlerView) Render(w http.ResponseWriter, req *http.Request, args []string) {
	v.Handler.ServeHTTP(w, req)
}

// A Resolver maps URLs to Views.
type Resolver interface {
	Resolve(string) (View, []string)
}

// A Pattern holds a single URL regular expression.
type Pattern struct {
	regexp *regexp.Regexp
	view   View
	name   string
}

func New(re string, view View, name string) *Pattern {
	return &Pattern{
		regexp: regexp.MustCompile(re),
		name:   name,
		view:   view,
	}
}

func (pat *Pattern) Resolve(p string) (v View, args []string) {
	m := pat.regexp.FindStringSubmatch(p)
	if m != nil {
		v, args = pat.view, m[1:]
	}
	return
}

type RegexResolver struct {
	regexp    *regexp.Regexp
	resolvers []Resolver
}

func Patterns(prefix string, resolvers ...Resolver) (r *RegexResolver) {
	return &RegexResolver{
		regexp:    regexp.MustCompile(prefix),
		resolvers: resolvers,
	}
}

func (resolver *RegexResolver) Resolve(p string) (v View, args []string) {
	m := resolver.regexp.FindStringSubmatchIndex(p)
	if m != nil {
		newPath := p[m[1]:]
		for _, r := range resolver.resolvers {
			var suffixArgs []string
			v, suffixArgs = r.Resolve(newPath)
			if v != nil {
				prefixArgCount := len(m)/2 - 1
				args = make([]string, prefixArgCount+len(suffixArgs))
				for i := 0; i < prefixArgCount; i++ {
					args[i] = p[m[(i+1)*2]:m[(i+1)*2+1]]
				}
				copy(args[prefixArgCount:], suffixArgs)
				return
			}
		}
	}
	return nil, nil
}

func clean(p string) (np string) {
	np = path.Clean("/" + p)[1:]
	if p[len(p)-1] == '/' && np != "" {
		np += "/"
	}
	return
}

func (resolver *RegexResolver) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	v, args := resolver.Resolve(clean(req.URL.Path))
	if v == nil {
		http.NotFound(w, req)
		return
	}
	v.Render(w, req, args)
}

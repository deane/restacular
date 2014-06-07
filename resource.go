package restacular

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Resource struct {
	basePath string
	routes   []*Route
}

type Handle func(ResponseWriter, *http.Request, map[string]string)

type Route struct {
	method  string
	pattern string
	handler httprouter.Handle
}

type ResourceHandler interface {
	Define() *Resource
}

func NewResource(basePath string) *Resource {
	return &Resource{basePath, []*Route{}}
}

func adaptHandler(handle Handle) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, params map[string]string) {
		writer := NewResponse(w)
		handle(writer, req, params)
	}
}

// AddRoute calls the HTTP_METHOD func or panic
func (resource *Resource) AddRoute(method string, pattern string, handler Handle) {
	methods := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"PATCH":   true,
		"DELETE":  true,
		"OPTIONS": true,
	}

	if _, ok := methods[method]; ok == false {
		log.Panicln("Tried to add an handler with a method that does not exist")
	}
	resource.routes = append(resource.routes, &Route{method, resource.basePath + pattern, adaptHandler(handler)})
}

func (resource *Resource) GET(pattern string, handler Handle) {
	resource.AddRoute("GET", pattern, handler)
}

func (resource *Resource) POST(pattern string, handler Handle) {
	resource.AddRoute("POST", pattern, handler)
}

func (resource *Resource) PUT(pattern string, handler Handle) {
	resource.AddRoute("PUT", pattern, handler)
}

func (resource *Resource) PATCH(pattern string, handler Handle) {
	resource.AddRoute("PATCH", pattern, handler)
}

func (resource *Resource) DELETE(pattern string, handler Handle) {
	resource.AddRoute("DELETE", pattern, handler)
}

func (resource *Resource) OPTIONS(pattern string, handler Handle) {
	resource.AddRoute("OPTIONS", pattern, handler)
}

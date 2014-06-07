package restacular

import (
	"github.com/julienschmidt/httprouter"
	"log"
)

type Resource struct {
	basePath string
	routes   []*Route
}

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

// AddRoute calls the HTTP_METHOD func or panic
func (resource *Resource) AddRoute(method string, pattern string, handler httprouter.Handle) {
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
	resource.routes = append(resource.routes, &Route{method, resource.basePath + pattern, handler})
}

func (resource *Resource) GET(pattern string, handler httprouter.Handle) {
	resource.AddRoute("GET", pattern, handler)
}

func (resource *Resource) POST(pattern string, handler httprouter.Handle) {
	resource.AddRoute("POST", pattern, handler)
}

func (resource *Resource) PUT(pattern string, handler httprouter.Handle) {
	resource.AddRoute("PUT", pattern, handler)
}

func (resource *Resource) PATCH(pattern string, handler httprouter.Handle) {
	resource.AddRoute("PATCH", pattern, handler)
}

func (resource *Resource) DELETE(pattern string, handler httprouter.Handle) {
	resource.AddRoute("DELETE", pattern, handler)
}

func (resource *Resource) OPTIONS(pattern string, handler httprouter.Handle) {
	resource.AddRoute("OPTIONS", pattern, handler)
}

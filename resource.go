package restacular

import (
	"log"
	"net/http"
)

type Resource struct {
	name     string
	basePath string
	routes   []*Route
}

type Route struct {
	method  string
	pattern string
	handler http.Handler
}

type ResourceHandler interface {
	DefineResource() Resource
}

func NewResource(name string, basePath string) *Resource {
	return &Resource{name, basePath, []*Route{}}
}

// AddRoute calls the HTTP_METHOD func or panic
func (resource *Resource) AddRoute(method string, pattern string, handler http.Handler) {
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
	resource.routes = append(resource.routes, &Route{method, resource.pattern + pattern, handler})
}

func (resource *Resource) GET(pattern string, handler http.Handler) {
	resource.AddRoute("GET", pattern, handler)
}

func (resource *Resource) POST(pattern string, handler http.Handler) {
	resource.AddRoute("POST", pattern, handler)
}

func (resource *Resource) PUT(pattern string, handler http.Handler) {
	resource.AddRoute("PUT", pattern, handler)
}

func (resource *Resource) PATCH(pattern string, handler http.Handler) {
	resource.AddRoute("PATCH", pattern, handler)
}

func (resource *Resource) DELETE(pattern string, handler http.Handler) {
	resource.AddRoute("DELETE", pattern, handler)
}

func (resource *Resource) OPTIONS(pattern string, handler http.Handler) {
	resource.AddRoute("OPTIONS", pattern, handler)
}

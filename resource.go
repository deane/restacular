package restacular

import (
	"net/http"
)

type Resource struct {
	name     string
	basePath string
}

type ResourceHandler interface {
	DefineResource() Resource
}

func NewResource(name string, basePath string) *Resource {
	return &Resource{name, basePath}
}

func (resource *Resource) AddRoute(method string, handler http.Handler) {
	// calls the HTTP_METHOD func or panic
}

func (resource *Resource) AddRouteWithPattern(method string, pattern string, handler http.Handler) {
	// calls the HTTP_METHOD func or panic
	// for stuff that doesn't map with a rest request
}

func (resource *Resource) GET(handler http.Handler) {

}

func (resource *Resource) POST(handler http.Handler) {

}

func (resource *Resource) PUT(handler http.Handler) {

}

func (resource *Resource) PATCH(handler http.Handler) {

}

func (resource *Resource) DELETE(handler http.Handler) {

}

func (resource *Resource) OPTIONS(handler http.Handler) {

}

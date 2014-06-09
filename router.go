package restacular

import (
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

// TODO: 404 handler to give httprouter

type Router struct {
	baseURL          string
	resourcesMapping map[string]string
	mux              *httprouter.Router
}

func NewRouter(baseURL string) *Router {
	return &Router{baseURL, make(map[string]string), httprouter.New()}
}

func (router *Router) AddResource(name string, resourceHandler ResourceHandler) {
	resource := resourceHandler.Define()

	if _, ok := router.resourcesMapping[name]; ok {
		log.Fatalf("A resource called %s is already registered.", name)
	}
	router.resourcesMapping[name] = resource.basePath

	for _, route := range resource.routes {
		router.mux.Handle(route.method, route.pattern, route.handler)
	}
}

func (router *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	router.mux.ServeHTTP(w, req)
}

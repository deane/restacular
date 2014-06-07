package restacular

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type Router struct {
	baseURL          string
	resourcesMapping map[string]string
	mux              *httprouter.Router
}

func NewRouter(baseURL string) *Router {
	return &Router{baseURL, make(map[string]string), httprouter.New()}
	// Need to give option to override stuff from httprouter
}

func (router *Router) AddResource(name string, resourceHandler ResourceHandler) {
	resource := resourceHandler.Define()
	router.resourcesMapping[name] = resource.basePath

	for _, route := range resource.routes {
		router.mux.Handle(route.method, route.pattern, route.handler)
	}
}

func (router *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	router.mux.ServeHTTP(w, req)
}

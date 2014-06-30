package restacular

import (
	"fmt"
	"net/http"
	"strings"
)

type router struct {
	baseURL          string
	resourcesMapping map[string]string
	tree             *node
	NotFoundHandler  func(http.ResponseWriter, *http.Request)
	PanicHandler     func(http.ResponseWriter, *http.Request, interface{})
}

func NewRouter(baseURL string) *router {
	return &router{
		baseURL:          baseURL,
		resourcesMapping: make(map[string]string),
		tree:             &node{path: "/"},
	}
}

type HandlerFunc func(Context, http.ResponseWriter, *http.Request)

type Context struct {
	Params map[string]string
	Env    map[string]interface{}
}

func (router *router) Handle(method string, path string, handler HandlerFunc) {
	// TODO: check method is valid
	// TODO: check that first char is / otherwise panic

	// Add a trailing slash to easily handle all cases
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	node := router.tree.addPath(strings.ToLower(path[1:]), true)

	node.setHandler(method, handler)
}

// GET is a shortcut for router.Handle("GET", path, handle)
func (router *router) GET(path string, handler HandlerFunc) {
	router.Handle("GET", path, handler)
}

// POST is a shortcut for router.Handle("POST", path, handle)
func (router *router) POST(path string, handler HandlerFunc) {
	router.Handle("POST", path, handler)
}

// PUT is a shortcut for router.Handle("PUT", path, handle)
func (router *router) PUT(path string, handler HandlerFunc) {
	router.Handle("PUT", path, handler)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle)
func (router *router) PATCH(path string, handler HandlerFunc) {
	router.Handle("PATCH", path, handler)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle)
func (router *router) DELETE(path string, handler HandlerFunc) {
	router.Handle("DELETE", path, handler)
}

func (router *router) AddResource(name string, resourceHandler ResourceHandler) {
	resource := resourceHandler.Define()

	if _, ok := router.resourcesMapping[name]; ok {
		panic(fmt.Sprintf("A resource called %s is already registered.", name))
	}
	router.resourcesMapping[name] = resource.basePath

	for _, route := range resource.routes {
		router.Handle(route.method, route.pattern, route.handler)
	}
}

func (router *router) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if router.PanicHandler != nil {
		defer func(resp http.ResponseWriter, req *http.Request) {
			if rcv := recover(); rcv != nil {
				router.PanicHandler(resp, req, rcv)
			}
		}(resp, req)
	}

	var params map[string]string
	path := req.URL.Path

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	node := router.tree.find(strings.ToLower(path[1:]), &params)

	if node != nil {
		if handler, ok := node.handlers[req.Method]; ok {
			context := Context{
				Params: params,
			}
			handler(context, resp, req)
			return
		}
	}

	if router.NotFoundHandler != nil {
		router.NotFoundHandler(resp, req)
	} else {
		http.NotFound(resp, req)
	}

}
func (r *router) PrintRoutes() {
	fmt.Println(r.tree.printTree("", ""))
}

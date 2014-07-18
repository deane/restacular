package restacular

import (
	"fmt"
	"net/http"
	"strings"
)

type router struct {
	baseURL                 string
	resourcesMapping        map[string]string
	tree                    *node
	NotFoundHandler         func(ResponseWriter, *http.Request)
	MethodNotAllowedHandler func(ResponseWriter, *http.Request, map[string]HandlerFunc)
	PanicHandler            func(ResponseWriter, *http.Request, interface{})
}

func NewRouter(baseURL string) *router {
	return &router{
		baseURL:          baseURL,
		resourcesMapping: make(map[string]string),
		tree:             &node{path: "/"},
	}
}

type HandlerFunc func(ResponseWriter, *http.Request)

func (router *router) Handle(method string, path string, handler HandlerFunc) {
	if path[0] != '/' {
		panic(fmt.Sprintf("Path %s must start with /", path))
	}
	// Remove leading slash first
	path = path[1:]
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	node := router.tree.addPath(path)
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

// AddResource adds a resource to the router, it will add all the routes it contains
// and panic if a resource with the same name is already registered
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
				router.PanicHandler(newResponse(resp, nil), req, rcv)
			}
		}(resp, req)
	}

	node, params := router.tree.find(req.URL.Path)

	if node != nil {
		if handler, ok := node.handlers[req.Method]; ok {
			handler(newResponse(resp, params), req)
			return
		} else {
			// 405
			if router.MethodNotAllowedHandler != nil {
				router.MethodNotAllowedHandler(newResponse(resp, params), req, node.handlers)
			} else {
				notAllowedHandler(newResponse(resp, params), req, node.handlers)
			}
			return
		}
	}

	// 404
	if router.NotFoundHandler != nil {
		router.NotFoundHandler(newResponse(resp, params), req)
	} else {
		http.NotFound(resp, req)
	}

}

// PrintRoutes is a debug utility that prints the trie with the priority/handlers to the console
func (r *router) PrintRoutes() {
	fmt.Println(r.tree.printTree("", ""))
}

// notAllowedHandler is a default handler for a 405 error, it sets the error code and the Allow header
// with the appropriate methods
func notAllowedHandler(resp ResponseWriter, req *http.Request, handlers map[string]HandlerFunc) {
	var methods []string

	for method := range handlers {
		methods = append(methods, method)
	}

	allowHeader := strings.Join(methods, ", ")

	if len(allowHeader) > 0 {
		resp.Header().Add("Allow", allowHeader)
	}
	resp.WriteHeader(http.StatusMethodNotAllowed)
}

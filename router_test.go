package restacular

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type PostResource struct{}

func (p PostResource) Define() *Resource {
	resource := NewResource("/posts")

	resource.GET("", p.List)
	resource.POST("/:post_id", p.Post)

	return resource
}

func (p PostResource) List(resp http.ResponseWriter, req *http.Request, params map[string]string) {
	resp.WriteHeader(200)
}

func (p PostResource) Post(resp http.ResponseWriter, req *http.Request, params map[string]string) {
	resp.WriteHeader(200)
}

// Small util function to quickly HTTP requests
func doRequest(method string, path string, router *Router, t *testing.T) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		t.Fatal("Errored when doing an HTTP request")
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestAddingResource(t *testing.T) {
	router := NewRouter("https://www.testing.com/api/")
	router.AddResource("posts", &PostResource{})

	w := doRequest("GET", "/posts", router, t)
	if w.Code != 200 {
		t.Errorf("Received http code %d instead of 200", w.Code)
	}

	w = doRequest("POST", "/posts/1", router, t)
	if w.Code != 200 {
		t.Errorf("Received http code %d instead of 200", w.Code)
	}

	// Testing that we get 404 when calling an existing path but with the wrong method
	w = doRequest("POST", "/posts", router, t)
	if w.Code != 404 {
		t.Errorf("Received http code %d instead of 404", w.Code)
	}
}

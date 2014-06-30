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

func (p PostResource) List(context Context, resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(200)
}

func (p PostResource) Post(context Context, resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(200)
}

// Small util function to quickly HTTP requests
func doRequest(method string, path string, router *router, t *testing.T) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		t.Fatal("Errored when doing an HTTP request")
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func createRouter() *router {
	router := NewRouter("https://www.testing.com/api/")
	router.AddResource("posts", &PostResource{})
	return router
}

func TestAddingResource(t *testing.T) {
	router := createRouter()
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

func TestFindingRouteCaseSensitivy(t *testing.T) {
	router := createRouter()
	w := doRequest("GET", "/POSTs", router, t)
	if w.Code != 200 {
		t.Errorf("Received http code %d instead of 200", w.Code)
	}
}

func TestFindingRouteTrailingSlash(t *testing.T) {
	router := createRouter()

	w := doRequest("GET", "/posts/", router, t)
	if w.Code != 200 {
		t.Errorf("Received http code %d instead of 200", w.Code)
	}

	w = doRequest("GET", "/posts", router, t)
	if w.Code != 200 {
		t.Errorf("Received http code %d instead of 200", w.Code)
	}
}

type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() (h http.Header) {
	return http.Header{}
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}

func (m *mockResponseWriter) WriteHeader(int) {}

func benchRequest(b *testing.B, router http.Handler, r *http.Request) {
	w := new(mockResponseWriter)
	u := r.URL
	rq := u.RawQuery
	r.RequestURI = u.RequestURI()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		u.RawQuery = rq
		router.ServeHTTP(w, r)
	}
}

func BenchmarkGettingRouteWithoutParam(b *testing.B) {
	router := createRouter()
	req, _ := http.NewRequest("GET", "/posts", nil)
	benchRequest(b, router, req)
}

func BenchmarkGettingRouteWithParam(b *testing.B) {
	router := createRouter()
	req, _ := http.NewRequest("GET", "/posts/1", nil)
	benchRequest(b, router, req)
}

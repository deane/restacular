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

func fakeView(context Context, resp http.ResponseWriter, req *http.Request) {}

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
	router := NewRouter("https://www.glass.com/")

	// Google glass API (https://developers.google.com/glass/v1/reference/)

	router.GET("/timeline", fakeView)
	router.GET("/timeline/:itemId", fakeView)
	router.PUT("/timeline/:itemId", fakeView)
	router.PATCH("/timeline/:itemId", fakeView)
	router.POST("/timeline", fakeView)
	router.DELETE("/timeline/:itemId", fakeView)

	router.GET("/timeline/:itemId/attachments/:attachmentId", fakeView)
	router.GET("/timeline/:itemId/attachments", fakeView)
	router.POST("/timeline/:itemId/attachments", fakeView)
	router.DELETE("/timeline/:itemId/attachments/:attachmentId", fakeView)

	router.GET("/subscriptions", fakeView)
	router.PUT("/subscriptions/:id", fakeView)
	router.POST("/subscriptions", fakeView)
	router.DELETE("/subscriptions/:id", fakeView)

	router.GET("/locations/:id", fakeView)
	router.GET("/locations", fakeView)

	router.GET("/contacts", fakeView)
	router.GET("/contacts/:id", fakeView)
	router.PUT("/contacts/:id", fakeView)
	router.PATCH("/contacts/:id", fakeView)
	router.POST("/contacts", fakeView)
	router.DELETE("/contacts/:id", fakeView)

	router.POST("/accounts/:userToken/:accountType/:accountName", fakeView)

	router.GET("/settings/:id", fakeView)

	return router
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

func TestFindingRouteCaseSensitivy(t *testing.T) {
	router := createRouter()
	w := doRequest("GET", "/timelINE", router, t)
	if w.Code != 200 {
		t.Errorf("Received http code %d instead of 200", w.Code)
	}
}

func TestFindingRouteTrailingSlash(t *testing.T) {
	router := createRouter()

	w := doRequest("GET", "/timeline", router, t)
	if w.Code != 200 {
		t.Errorf("Received http code %d instead of 200", w.Code)
	}

	w = doRequest("GET", "/timeline/", router, t)
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
	req, _ := http.NewRequest("GET", "/timeline", nil)
	benchRequest(b, router, req)
}

func BenchmarkGettingRouteWithOneParam(b *testing.B) {
	router := createRouter()
	req, _ := http.NewRequest("GET", "/timeline/:itemId", nil)
	benchRequest(b, router, req)
}

func BenchmarkGettingRouteWithTwoParam(b *testing.B) {
	router := createRouter()
	req, _ := http.NewRequest("GET", "/timeline/:itemId/attachments/:attachmentId", nil)
	benchRequest(b, router, req)
}

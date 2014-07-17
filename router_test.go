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
	resource.POST("/:postId/", p.Post)

	return resource
}

func (p PostResource) List(context Context, resp ResponseWriter, req *http.Request) {
	resp.WriteHeader(200)
}

func (p PostResource) Post(context Context, resp ResponseWriter, req *http.Request) {
	resp.WriteHeader(200)
}

func fakeView(context Context, resp ResponseWriter, req *http.Request) {}

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

	// Testing that we get 405 when calling an existing path but with the wrong method
	w = doRequest("POST", "/posts", router, t)
	if w.Code != 405 {
		t.Errorf("Received http code %d instead of 405", w.Code)
	}
}

func TestAddingDuplicateResource(t *testing.T) {
	router := NewRouter("https://www.testing.com/api/")
	panicked := false
	panicHandler := func() {
		if err := recover(); err != nil {
			panicked = true
		}
	}

	addPanicResource := func(resource ResourceHandler) {
		panicked = false
		defer panicHandler()
		router.AddResource("blob", resource)
	}
	addPanicResource(&PostResource{})
	addPanicResource(&PostResource{})
	if !panicked {
		t.Errorf("Adding several resources with the same should have panicked")
	}
}

type testRoute struct {
	method       string
	path         string
	expectedCode int
}

func TestFindRoutes(t *testing.T) {
	router := createRouter()

	requests := []testRoute{
		{"GET", "/timeline", 200},
		{"GET", "/timeLINE", 200},                 // case sensitivity
		{"GET", "/timeline/", 200},                // trailing slash
		{"GET", "/students", 404},                 // 404
		{"DELETE", "/timeline", 405},              // no DELETE on that url
		{"GET", "/timeline/1/attachments/2", 200}, // 404
		{"DELETE", "/subscriptions/21", 200},
		{"PUT", "/subscriptions/21", 200},
		{"POST", "/subscriptions", 200},
		{"PATCH", "/contacts/21", 200},
	}

	for _, request := range requests {
		w := doRequest(request.method, request.path, router, t)
		if w.Code != request.expectedCode {
			t.Errorf("Received http code %d instead of %d for %s [%s]", w.Code, request.expectedCode, request.path, request.method)
		}
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
